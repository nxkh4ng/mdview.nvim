local M = {}
local server = require("mdview.server")
local autocmds = require("mdview.autocmds")

function M.start()
	if vim.bo.filetype ~= "markdown" then
		vim.notify("[mdview] not in markdown buffer", vim.log.levels.ERROR)
		return
	end

	server.start()
	autocmds.enable()
end

function M.stop()
	server.stop()
	autocmds.disable()
end

function M.setup()
	vim.api.nvim_create_user_command("MdviewStart", M.start, {})
	vim.api.nvim_create_user_command("MdviewStop", M.stop, {})
end

return M
