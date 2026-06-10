local M = {}

local defaults = {
	host = "127.0.0.1",
	port = 0,
	browser = "",
}

function M.setup(opts)
	if opts then
		defaults = vim.tbl_deep_extend("force", defaults, opts)
	end
end

function M.get()
	return defaults
end

return M
