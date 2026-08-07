local M = {}

local GITHUB_OWNER = "nxkh4ng"
local GITHUB_REPO = "mdview.nvim"

local sep = vim.fn.has("win32") == 1 and "\\" or "/"
local INSTALL_DIR = vim.fn.stdpath("data") .. sep .. "mdview.nvim"
local VERSION_FILE = INSTALL_DIR .. sep .. "version"
local LOCK_FILE = INSTALL_DIR .. sep .. ".lock"

local function normalize_arch(arch)
	arch = arch:lower()
	if arch == "x86_64" or arch == "amd64" then
		return "amd64"
	elseif arch == "aarch64" or arch == "arm64" then
		return "arm64"
	end
	return nil
end

local function detect_platform()
	local os_name = vim.uv.os_uname().sysname
	local arch = normalize_arch(vim.uv.os_uname().machine)
	if not arch then
		return nil
	end
	if os_name == "Linux" then
		return (arch == "amd64") and "linux-amd64" or "linux-arm64"
	elseif os_name == "Darwin" then
		return (arch == "amd64") and "darwin-amd64" or "darwin-arm64"
	elseif os_name == "Windows_NT" then
		return (arch == "amd64") and "windows-amd64" or "windows-arm64"
	end
	return nil
end

local os_platform = detect_platform()
if not os_platform then
	vim.notify(
		string.format(
			"[mdview] unsupported platform: %s/%s",
			vim.uv.os_uname().sysname,
			vim.uv.os_uname().machine
		),
		vim.log.levels.ERROR
	)
	return M
end

local is_windows = vim.fn.has("win32") == 1
local archive_ext = is_windows and ".zip" or ".tar.gz"
local ARCHIVE_NAME = "mdview-" .. os_platform .. archive_ext
local BINARY_NAME = is_windows and "mdview.exe" or "mdview"
local BINARY_PATH = INSTALL_DIR .. sep .. BINARY_NAME
M.binary_path = BINARY_PATH

local function notify(msg, level)
	vim.notify("[mdview] " .. msg, level or vim.log.levels.INFO)
end

local function run(cmd, timeout_ms)
	local out, code
	vim.system(cmd, { text = true }, function(obj)
		out = obj.stdout or ""
		code = obj.code
	end)
	local ok = vim.wait(timeout_ms or 120000, function()
		return out ~= nil
	end)
	if not ok then
		return nil, "timeout"
	end
	return out, code == 0
end

local function download(url, dest)
	if vim.fn.executable("curl") == 1 then
		return run({ "curl", "-fL", "-sS", "-o", dest, url })
	elseif vim.fn.executable("wget") == 1 then
		return run({ "wget", "-q", "-O", dest, url })
	end
	return nil, "curl or wget required"
end

local function sha256_file(path)
	local out, ok
	if is_windows then
		out, ok = run({
			"powershell",
			"-NoProfile",
			"-Command",
			"(Get-FileHash -Algorithm SHA256 -Path '"
				.. path
				.. "').Hash.ToLower()",
		})
	elseif vim.fn.has("mac") == 1 then
		out, ok = run({ "shasum", "-a", "256", path })
	else
		out, ok = run({ "sha256sum", path })
	end
	if not ok or not out then
		return nil
	end
	local h = out:match("^%s*([0-9a-fA-F]+)")
	return h and h:lower() or nil
end

local function verify_checksum(archive_path, tag)
	local ver = tag:gsub("^[Vv]", "")
	local checksum_name = "mdview_" .. ver .. "_checksums.txt"
	local url = string.format(
		"https://github.com/%s/%s/releases/download/%s/%s",
		GITHUB_OWNER,
		GITHUB_REPO,
		tag,
		checksum_name
	)

	local tmp = INSTALL_DIR .. sep .. "checksums.txt"
	local _, ok = download(url, tmp)
	if not ok then
		return false, "cannot download checksums.txt"
	end
	local f = io.open(tmp, "r")
	if not f then
		return false, "cannot read checksums.txt"
	end
	local want
	for line in f:lines() do
		local hash, name = line:match("^([%x]+)%s+(.+)$")
		if name and name:gsub("/", "") == ARCHIVE_NAME then
			want = hash:lower()
		end
	end
	f:close()
	os.remove(tmp)
	if not want then
		return false, "checksum not found for " .. ARCHIVE_NAME
	end
	if sha256_file(archive_path) ~= want then
		return false, "checksum mismatch"
	end
	return true
end

local function git_tag(dir)
	if not dir or not vim.uv.fs_stat(dir) then
		return nil
	end
	local out, ok =
		run({ "git", "-C", dir, "describe", "--tags", "--abbrev=0" }, 15000)
	if not ok or not out then
		return nil
	end
	local tag = out:match("^%s*(.-)%s*$")
	return (tag ~= "" and tag) or nil
end

local function latest_tag()
	local url = string.format(
		"https://api.github.com/repos/%s/%s/releases/latest",
		GITHUB_OWNER,
		GITHUB_REPO
	)
	local tmp = INSTALL_DIR .. sep .. "latest.json"
	local _, ok = download(url, tmp)
	if not ok then
		return nil, "cannot query latest release"
	end
	local f = io.open(tmp, "r")
	if not f then
		return nil, "cannot read latest.json"
	end
	local data = f:read("*a")
	f:close()
	os.remove(tmp)
	local ok2, json = pcall(vim.json.decode, data)
	if not ok2 or not json or not json.tag_name then
		return nil, "invalid latest response"
	end
	return json.tag_name
end

local function resolve_tag(dir)
	local t = git_tag(dir)
	if t then
		return t
	end
	return latest_tag()
end

local function read_version()
	local f = io.open(VERSION_FILE, "r")
	if not f then
		return nil
	end
	local v = f:read("*l")
	f:close()
	return v and v:match("^%s*(.-)%s*$") or nil
end

local function write_version(tag)
	local f = io.open(VERSION_FILE, "w")
	if f then
		f:write(tag, "\n")
		f:close()
	end
end

local function acquire_lock(timeout_ms)
	vim.fn.mkdir(INSTALL_DIR, "p")
	local deadline = vim.loop.hrtime() + (timeout_ms or 120000) * 1e6
	while true do
		local fd = vim.uv.fs_open(LOCK_FILE, "wx", 420)
		if fd then
			vim.uv.fs_close(fd)
			return true
		end
		local stat = vim.uv.fs_stat(LOCK_FILE)
		if stat then
			local age_ms = (
				vim.loop.hrtime() - (stat.mtime.sec * 1e9 + stat.mtime.nsec)
			) / 1e6
			if age_ms > 600000 then
				vim.uv.fs_unlink(LOCK_FILE)
				vim.wait(100)
			end
		end
		if vim.loop.hrtime() > deadline then
			return false
		end
		vim.wait(200)
	end
end

local function release_lock()
	vim.uv.fs_unlink(LOCK_FILE)
end

function M.install(dir)
	if not os_platform then
		return M
	end
	if not acquire_lock() then
		notify("another install in progress", vim.log.levels.ERROR)
		return M
	end

	local tag, err = resolve_tag(dir)
	if not tag then
		release_lock()
		notify(
			"cannot resolve release version: " .. tostring(err),
			vim.log.levels.ERROR
		)
		return M
	end

	local archive = INSTALL_DIR .. sep .. ARCHIVE_NAME
	local url = string.format(
		"https://github.com/%s/%s/releases/download/%s/%s",
		GITHUB_OWNER,
		GITHUB_REPO,
		tag,
		ARCHIVE_NAME
	)

	local _, ok = download(url, archive)
	if not ok then
		release_lock()
		notify("download failed: " .. url, vim.log.levels.ERROR)
		return M
	end

	local vok, verr = verify_checksum(archive, tag)
	if not vok then
		os.remove(archive)
		release_lock()
		notify(
			"checksum verify failed: " .. tostring(verr),
			vim.log.levels.ERROR
		)
		return M
	end

	if is_windows then
		run({ "tar", "xf", archive, "-C", INSTALL_DIR })
	else
		run({ "tar", "xzf", archive, "-C", INSTALL_DIR })
	end
	os.remove(archive)

	if not is_windows then
		run({ "chmod", "+x", BINARY_PATH })
	end

	write_version(tag)
	release_lock()

	local stat = vim.uv.fs_stat(BINARY_PATH)
	notify(
		"installed "
			.. BINARY_NAME
			.. " "
			.. tag
			.. (stat and (" (" .. stat.size .. " bytes)") or "")
	)
	return M
end

local function plugin_root()
	local path = (debug.getinfo(1, "S").source or ""):gsub("^@", "")
	if path == "" then
		return nil
	end
	return vim.fn.fnamemodify(path, ":h:h:h")
end

function M.ensure()
	local dir = plugin_root()
	if not dir or not vim.uv.fs_stat(dir) then
		notify("cannot locate plugin directory", vim.log.levels.ERROR)
		return
	end
	if vim.uv.fs_stat(BINARY_PATH) then
		local want = git_tag(dir)
		if not want then
			return
		end
		if read_version() == want then
			return
		end
	end
	M.install(dir)
end

return M
