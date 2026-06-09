local M = {}
local server = require("mdview.server")

function M.start()
	server.start()
end
function M.stop()
	server.stop()
end
function M.refresh()
	server.refresh()
end

function M.setup()
	vim.api.nvim_create_user_command("MdviewStart", M.start, {})
	vim.api.nvim_create_user_command("MdviewStop", M.stop, {})
	vim.api.nvim_create_user_command("MdviewRefresh", M.refresh, {})
end

return M
