local M = {}

local process = nil
local port = nil
local binary = vim.fn.stdpath("data") .. "/mdview.nvim/mdview-server"

function M.start(cfg_host, cfg_port, cfg_browser)
	if not vim.uv.fs_stat(binary) then
		vim.notify("[mdview] binary not found at " .. binary, vim.log.levels.ERROR)
		return
	end

	if process and not process:is_closing() then
		vim.notify("[mdview] server already running", vim.log.levels.WARN)
		return
	end

	local args = { binary }
	table.insert(args, "-host")
	table.insert(args, cfg_host)

	if cfg_port ~= 0 then
		table.insert(args, "-port")
		table.insert(args, tostring(cfg_port))
	end

	if cfg_browser ~= "" then
		table.insert(args, "-browser")
		table.insert(args, cfg_browser)
	end

	process = vim.system(args, {
		text = true,
		detach = true,
		stdout = function(_, data)
			if data and data ~= "" then
				port = tonumber(data:match("%d+"))
				if port then
					vim.schedule(function()
						vim.notify("[mdview] port=" .. port, vim.log.levels.INFO)
					end)
				end
			end
		end,
		stderr = function(_, data)
			if data and data ~= "" then
				vim.schedule(function()
					vim.notify("[mdview] " .. data, vim.log.levels.ERROR)
				end)
			end
		end,
	}, nil)
end

function M.stop()
	if not process or process:is_closing() then
		vim.notify("[mdview] server not running", vim.log.levels.WARN)
		return
	end

	process:kill("sigterm")
	process = nil
	port = nil

	vim.notify("[mdview] server stopped", vim.log.levels.INFO)
end

function M.get_port()
	return port
end

return M
