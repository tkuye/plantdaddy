### Author: Tosin Kuye
### This is the script to run the majority of the microcontrollers code. This
### It has three components: The dht, light and moisture sensor parts.

from boot import get_configs
import machine
from machine import ADC, Pin
import dht
import time
import urequests
import ujson

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
		write_to_config()
	elif voltage_read < wet_max:
		CONFIGURATIONS["wet_max"] = voltage_read
		write_to_config()
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

def write_to_config():
	"""Writes a key value pair to the config file.

	Args:
		key (str): The key to write to the config file.
		value (Any): The value to write to the config file.
	"""
	global CONFIGURATIONS
	with open("config.json", "w") as f:
		ujson.dump(CONFIGURATIONS, f)


def get_session_id():
	"""
	Creates an http response to get the sessionID from the server
	"""
	global CONFIGURATIONS

	device_id = CONFIGURATIONS.get("deviceID")
	url:str = CONFIGURATIONS.get("url")

	device_obj = ujson.dumps({"deviceID": device_id})
	res = post_data("/auth-device", device_obj)
	for key, val in res.json().items():
		CONFIGURATIONS[key] = val
	write_to_config()

def post_data(path: str, data: str) -> urequests.Response:
	"""
	Encloses the data for simple post requests.

	Args:
		path (str): The path to the data.
		data (str): The data to send to the server.

	Returns:
		urequests.Response: [description]
	"""
	global CONFIGURATIONS
	url = CONFIGURATIONS.get("url")
	res = urequests.post(url + path, 
	headers={'content-type':'application/json'},
	data=data)
	return res


def send_plant_data(temperature: int, humidity: int, soil_moisture: float, light: int):
	"""Send the plant data to the server

	Args:
		temperature (int)
		humidity (int)
		soil_moisture (float)
		light (int)
	"""
	global CONFIGURATIONS
	url = CONFIGURATIONS.get("url")
	session_id = CONFIGURATIONS.get("sessionID")
	usage_counter = CONFIGURATIONS.get("usageCounter")
	timestamp = CONFIGURATIONS.get("timestamp")
	device_id = CONFIGURATIONS.get("deviceID")
	data = {"sessionID":session_id,
	"usageCounter":usage_counter,
	"timestamp": timestamp,
	"temperature":temperature,
	"humidity":humidity,
	"soilMoisture":soil_moisture,
	"light":light, 
	"deviceID": device_id}

	json_data = ujson.dumps(data)

	res = post_data("/new-data", json_data)
	for key, value in res.json().items():
		CONFIGURATIONS[key] = value
	write_to_config()

def main():
	"""
	Runs the main function for the program.
	"""

	# Assume it is not none
	global CONFIGURATIONS
	sessionID = CONFIGURATIONS.get("sessionID")

	# If is is none now we send our request to get one.
	if sessionID is None:
		get_session_id()
	


	try:
		light = get_light_level()
		temperature, humidity = get_temperature()
		soil_moisture = get_moisture_level()
		send_plant_data(temperature, humidity, soil_moisture, light)
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