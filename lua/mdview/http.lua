local M = {}
local server = require("mdview.server")
local config = require("mdview.config")

function M.post(path, data)
	local port = server.get_port()
	if not port then
		vim.notify("[mdview] server port not ready yet. Try again", vim.log.levels.ERROR)
		return false
	end

	local cfg = config.get()
	local addr = cfg.host .. ":" .. port

	local body = vim.fn.json_encode(data)
	local request = table.concat({
		"POST " .. path .. " HTTP/1.1",
		"Host: " .. cfg.host .. ":" .. port,
		"Content-Type: application/json",
		"Content-Length: " .. #body,
		"",
		body,
	}, "\r\n")

	local ok, ch = pcall(vim.fn.sockconnect, "tcp", addr)
	if not ok then
		vim.notify("[mdview] connection failed - " .. ch, vim.log.levels.ERROR)
		return false
	end
	if ch == 0 then
		vim.notify("[mdview] connection refused", vim.log.levels.ERROR)
		return false
	end

	vim.fn.chansend(ch, request)
	return true
end

return M
