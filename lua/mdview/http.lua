local M = {}
local server = require("mdview.server")
local config = require("mdview.config")

local persistent_ch = nil
local persistent_addr = nil

local function connect(addr)
	local ok, ch = pcall(vim.fn.sockconnect, "tcp", addr, {
		on_data = function(_, _) end,
	})
	if not ok or ch == 0 then
		return nil
	end

	return ch
end

function M.post(path, data)
	local port = server.get_port()
	if not port then
		vim.notify("[mdview] server port not ready yet. Try again", vim.log.levels.ERROR)
		return false
	end

	local cfg = config.get()
	local addr = cfg.host .. ":" .. port

	if persistent_ch == nil or addr ~= persistent_addr then
		if persistent_ch then
			pcall(vim.fn.chanclose, persistent_ch)
		end
		persistent_ch = connect(addr)
		persistent_addr = addr

		if not persistent_ch then
			vim.notify("[mdview] cannot connect to server", vim.log.levels.ERROR)
			return false
		end
	end

	local body = vim.fn.json_encode(data)
	local request = table.concat({
		"POST " .. path .. " HTTP/1.1",
		"Host: " .. addr,
		"Content-Type: application/json",
		"Content-Length: " .. #body,
		"",
		body,
	}, "\r\n")

	local ok, _ = pcall(vim.fn.chansend, persistent_ch, request)
	if not ok then
		return false
	end

	return true
end

function M.close()
	if persistent_ch then
		pcall(vim.fn.chanclose, persistent_ch)
		persistent_ch = nil
		persistent_addr = nil
	end
end

return M
