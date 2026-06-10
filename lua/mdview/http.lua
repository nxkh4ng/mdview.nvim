local M = {}
local server = require("mdview.server")

function M.post(path, data)
	local port = server.get_port()
	if not port then
		vim.notify("[mdview] server port not ready yet. Try again", vim.log.levels.ERROR)
		return false
	end

	local body = vim.fn.json_encode(data)
	local request = table.concat({
		"POST " .. path .. " HTTP/1.1",
		"Host: 127.0.0.1:" .. port,
		"Content-Type: application/json",
		"Content-Length: " .. #body,
		"",
		body,
	}, "\r\n")

	local ok, ch = pcall(vim.fn.sockconnect, "tcp", "127.0.0.1:" .. port)
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
