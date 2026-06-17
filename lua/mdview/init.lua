local M = {}
local autocmds = require("mdview.autocmds")
local config = require("mdview.config")
local server = require("mdview.server")

function M.start()
	if vim.bo.filetype ~= "markdown" then
		vim.notify("[mdview] not in markdown buffer", vim.log.levels.ERROR)
		return
	end

	server.start(config.get())
	autocmds.enable()
end

function M.stop()
	server.stop()
	autocmds.disable()
end

function M.setup(opts)
	config.setup(opts)

	vim.api.nvim_create_user_command("MdviewStart", M.start, {})
	vim.api.nvim_create_user_command("MdviewStop", M.stop, {})
end

return M
