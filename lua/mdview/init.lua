local M = {}
local server = require("mdview.server")
local autocmds = require("mdview.autocmds")

local config = {
	host = "127.0.0.1",
	port = 0,
	browser = "",
}

function M.start()
	if vim.bo.filetype ~= "markdown" then
		vim.notify("[mdview] not in markdown buffer", vim.log.levels.ERROR)
		return
	end

	server.start(config.host, config.port, config.browser)
	autocmds.enable()
end

function M.stop()
	server.stop()
	autocmds.disable()
end

function M.setup(opts)
	if opts then
		config = vim.tbl_deep_extend("force", config, opts)
	end

	vim.api.nvim_create_user_command("MdviewStart", M.start, {})
	vim.api.nvim_create_user_command("MdviewStop", M.stop, {})
end

return M
