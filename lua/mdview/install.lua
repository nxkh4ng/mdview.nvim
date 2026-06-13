local M = {}

local GITHUB_OWNER = "nxkh4ng"
local GITHUB_REPO = "mdview.nvim"

local sep = vim.fn.has("win32") == 1 and "\\" or "/"
local INSTALL_DIR = vim.fn.stdpath("data") .. sep .. "mdview.nvim"

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
		return (arch == "amd64") and "-linux-amd64" or "-linux-arm64"
	elseif os_name == "Darwin" then
		return (arch == "amd64") and "-darwin-amd64" or "-darwin-arm64"
	elseif os_name == "Windows_NT" then
		return (arch == "amd64") and "-windows-amd64.exe" or "-windows-arm64.exe"
	end

	return nil
end

local suffix = detect_platform()
if not suffix then
	local os_name = vim.uv.os_uname().sysname
	local arch = vim.uv.os_uname().machine

	local err_msg = string.format("[mdview] unsupported platform: %s/%s", os_name, arch)
	vim.notify(err_msg, vim.log.levels.ERROR)
	return M
end

local BINARY_NAME = "mdview" .. suffix
local BINARY_PATH = INSTALL_DIR .. sep .. BINARY_NAME
M.binary_path = BINARY_PATH

vim.fn.mkdir(INSTALL_DIR, "p")

if vim.uv.fs_stat(BINARY_PATH) then
	return M
end

local url =
	string.format("https://github.com/%s/%s/releases/latest/download/%s", GITHUB_OWNER, GITHUB_REPO, BINARY_NAME)

local ok
if vim.fn.executable("curl") == 1 then
	vim.fn.system({ "curl", "-fL", "-o", BINARY_PATH, url })
	ok = vim.v.shell_error == 0
elseif vim.fn.executable("wget") == 1 then
	vim.fn.system({ "wget", "-O", BINARY_PATH, url })
	ok = vim.v.shell_error == 0
else
	vim.notify("[mdview] curl or wget required to download binary", vim.log.levels.ERROR)
	return M
end

if not ok then
	vim.notify("[mdview] Download failed: " .. url, vim.log.levels.ERROR)
	return M
end

if vim.fn.has("win32") ~= 1 then
	vim.fn.system({ "chmod", "+x", BINARY_PATH })
end

local final_stat = vim.uv.fs_stat(BINARY_PATH)
if final_stat and final_stat.size > 0 then
	vim.notify("[mdview] Downloaded: " .. BINARY_NAME .. " - size: " .. final_stat.size, vim.log.levels.INFO)
else
	vim.notify("[mdview] Downloaded binary is invalid", vim.log.levels.ERROR)
end

return M
