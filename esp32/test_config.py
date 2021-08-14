"""Test the write_to_config functionality
"""
def write_to_config(ckey: str, cvalue):
	"""Writes a key value pair to the config file.

	Args:
		key (str): The key to write to the config file.
		value (Any): The value to write to the config file.
	"""
	with open("config.txt", "r+") as f:
		lines = []
		found_line = False
		for line in f.readlines():
			key = line.split("=")[0]
			if key == ckey:
				line = "{}={}\n".format(ckey, cvalue)
				found_line = True
			lines.append(line)
		if not found_line:
			lines.append("{}={}\n".format(ckey, cvalue))
		f.seek(0)
		lines = "".join(lines)
		f.write(lines)

write_to_config("dry_max", 696969)
write_to_config("Hello", "World")
