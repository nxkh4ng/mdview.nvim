local M = {}

local process = nil
local port = nil
local binary = vim.fn.stdpath("data") .. "/mdview.nvim/mdview-server"

function M.start()
	if not vim.uv.fs_stat(binary) then
		vim.schedule(function()
			vim.notify("[mdview] binary not found at " .. binary, vim.log.levels.ERROR)
		end)
		return
	end

	if process and not process:is_closing() then
		vim.schedule(function()
			vim.notify("[mdview] server already running", vim.log.levels.WARN)
		end)
		return
	end

	process = vim.system({ binary }, {
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
		vim.schedule(function()
			vim.notify("[mdview] server not running", vim.log.levels.WARN)
		end)
		return
	end

	process:kill("sigterm")
	process = nil
	port = nil

	vim.schedule(function()
		vim.notify("[mdview] server stopped", vim.log.levels.INFO)
	end)
end

function M.refresh()
	M.stop()
	M.start()
end

return M
