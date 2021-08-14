### Author: Tosin Kuye
### This is the script to run the majority of the microcontrollers code. This
### It has three components: The dht, light and moisture sensor parts.

from boot import get_configs
import machine
from machine import ADC, Pin
import dht
import time

MS_TO_SECS = 1000

CONFIGURATIONS = get_configs()

def get_temperature():
	"""Gets the temperature from the sensor and returns the data as a tuple
	"""
	d = dht.DHT11(Pin(4))
	d.measure()
	return (d.temperature(), d.humidity())


def get_moisture_level():
	"""
	Gets the moisture level from the soil. 
	
	Returns as a percentage of the max and min.
	"""

	moisture_adc = ADC(Pin(34))
	moisture_adc.atten(ADC.ATTN_11DB)
	voltage_read = moisture_adc.read_u16()
	# Get the dry and wet starting values from the config file
	global CONFIGURATIONS
	dry_max = int(CONFIGURATIONS.get("dry_max"))
	wet_max = int(CONFIGURATIONS.get("wet_max"))

	if voltage_read > dry_max:
		dry_max = voltage_read
		CONFIGURATIONS["dry_max"]  = voltage_read
		write_to_config("dry_max", voltage_read)
	elif voltage_read < wet_max:
		CONFIGURATIONS["wet_max"] = voltage_read
		write_to_config("wet_max", voltage_read)
	return get_percentage(dry_max, wet_max, voltage_read) * 100


def get_light_level():
	"""
	Returns the light level from the photoresistor.
	"""
	light_adc = ADC(Pin(33))
	#light_adc.atten(ADC.ATTN_0DB)
	light_adc.width(ADC.WIDTH_12BIT)
	voltage_read = light_adc.read_u16()
	return get_percentage(65535, 0, voltage_read) * 100


def get_percentage(maxi: int, mini: int, value: int):
	"""Gets the percentage between the given values

	Args:
		maxi (int): max value
		mini (int): min value
		value (int): value
	"""
	return ((maxi - value) / (maxi - mini))

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


def main():
	"""
	Runs the main function for the program.
	"""
	try:
		light = get_light_level()
		temperature, humidity = get_temperature()
		soil_moisture = get_moisture_level()
		print("Current Temperature: {:d}\u2103 \tCurrent Humidity:{:d}%\tCurrent Soil Moisture: {:>.3f}%\tCurrent Light: {:>.3f}%"
			.format(temperature,
			humidity,
			soil_moisture,
			light))
		# Send the device to sleep after retrieving the data
		# Sleep for 10 minutes
		# Must regular sleep so commands can be sent
		print("About to sleep for 10 minutes, there is 5 seconds to send commands.")
		time.sleep(5)
		print("Going to sleep...")
		machine.deepsleep(600*MS_TO_SECS)

	except RuntimeError:
		pass

if __name__ == "__main__":
	main()