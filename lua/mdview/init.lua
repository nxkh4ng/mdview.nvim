local M = {}

function M.start()
	vim.notify("mdview: starting...")
end
function M.stop()
	vim.notify("mdview: stopping...")
end
function M.refresh()
	vim.notify("mdview: refreshing...")
end

vim.api.nvim_create_user_command("MdviewStart", M.start, {})
vim.api.nvim_create_user_command("MdviewStop", M.stop, {})
vim.api.nvim_create_user_command("MdviewRefresh", M.refresh, {})

return M
