function HandleRequest(url)
	if url == "/d/queuenamp" then
		local id = server.queueCommand("nmap")
		return "<b>Task Id: </b>" .. id
	end
	if url == "/d/getqueue" then
		local output = server.getCommandOutput("gzftlMtW4x30bD4uzu94TRZfM5uwwbPA")
		if output == nil then
			return ""
		end
		return output
	end
	return "nix"
end