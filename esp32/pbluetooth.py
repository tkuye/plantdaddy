from main import write_to_config, CONFIGURATIONS
import ubluetooth as bluetooth
from micropython import const
import utime as time
import ujson
from helpers import get_configs, do_connect
import machine

_IRQ_CENTRAL_CONNECT = const(1)
_IRQ_CENTRAL_DISCONNECT = const(2)
_IRQ_GATTS_WRITE = const(3)



_UART_SERVICE_UUID = bluetooth.UUID("5f7937b4-039f-11ec-9a03-0242ac130003")
_UART_RX_CHAR_UUID = bluetooth.UUID("5f7937b4-039f-11ec-9a03-0242ac130003")
_UART_TX_CHAR_UUID = bluetooth.UUID("5f7937b4-039f-11ec-9a03-0242ac130003")

_UART_TX = (
	_UART_TX_CHAR_UUID,
	bluetooth.FLAG_READ | bluetooth.FLAG_NOTIFY
)

_UART_RX = (
	_UART_RX_CHAR_UUID,
	bluetooth.FLAG_WRITE
)
_UART_SERVICE = (
	_UART_SERVICE_UUID,
	(_UART_TX, _UART_RX)
)

import struct


# Advertising payloads are repeated packets of the following form:
#   1 byte data length (N + 1)
#   1 byte type (see constants below)
#   N bytes type-specific data

_ADV_TYPE_FLAGS = const(0x01)
_ADV_TYPE_NAME = const(0x09)
_ADV_TYPE_UUID16_COMPLETE = const(0x3)
_ADV_TYPE_UUID32_COMPLETE = const(0x5)
_ADV_TYPE_UUID128_COMPLETE = const(0x7)
_ADV_TYPE_UUID16_MORE = const(0x2)
_ADV_TYPE_UUID32_MORE = const(0x4)
_ADV_TYPE_UUID128_MORE = const(0x6)
_ADV_TYPE_APPEARANCE = const(0x19)


# Generate a payload to be passed to gap_advertise(adv_data=...).
def advertising_payload(limited_disc=False, br_edr=False, name=None, services=None, appearance=0):
    payload = bytearray()

    def _append(adv_type, value):
        nonlocal payload
        payload += struct.pack("BB", len(value) + 1, adv_type) + value

    _append(
        _ADV_TYPE_FLAGS,
        struct.pack("B", (0x01 if limited_disc else 0x02) + (0x18 if br_edr else 0x04)),
    )

    if name:
        _append(_ADV_TYPE_NAME, name)

    if services:
        for uuid in services:
            b = bytes(uuid)
            if len(b) == 2:
                _append(_ADV_TYPE_UUID16_COMPLETE, b)
            elif len(b) == 4:
                _append(_ADV_TYPE_UUID32_COMPLETE, b)
            elif len(b) == 16:
                _append(_ADV_TYPE_UUID128_COMPLETE, b)

    # See org.bluetooth.characteristic.gap.appearance.xml
    if appearance:
        _append(_ADV_TYPE_APPEARANCE, struct.pack("<h", appearance))

    return payload


def decode_field(payload, adv_type):
    i = 0
    result = []
    while i + 1 < len(payload):
        if payload[i + 1] == adv_type:
            result.append(payload[i + 2 : i + payload[i] + 1])
        i += 1 + payload[i]
    return result


def decode_name(payload):
    n = decode_field(payload, _ADV_TYPE_NAME)
    return str(n[0], "utf-8") if n else ""


def decode_services(payload):
    services = []
    for u in decode_field(payload, _ADV_TYPE_UUID16_COMPLETE):
        services.append(bluetooth.UUID(struct.unpack("<h", u)[0]))
    for u in decode_field(payload, _ADV_TYPE_UUID32_COMPLETE):
        services.append(bluetooth.UUID(struct.unpack("<d", u)[0]))
    for u in decode_field(payload, _ADV_TYPE_UUID128_COMPLETE):
        services.append(bluetooth.UUID(u))
    return services


class BlePeripheral:
	"""
	Allows for the bluetooth peripheral to receive data from the iphone application.
	"""
	def __init__(self, ble, name="PL_UART"):
		self._ble = ble
		self._ble.active(True)
		self._ble.irq(self._irq)
		((self._handle_tx, self._handle_rx),) = self._ble.gatts_register_services((_UART_SERVICE,))
		self._connections = set()
		self._write_callback = None
		self._payload = advertising_payload(name=name, services=[_UART_SERVICE_UUID])
		self._advertise()
		self.data = str()
	
	def _advertise(self, interval_us=100000):
		print("Starting advertising")
		self._ble.gap_advertise(interval_us, adv_data=self._payload)

	def _irq(self, event, data):
		if event == _IRQ_CENTRAL_CONNECT:
			conn_handle, _, _ = data
			print("New connection", conn_handle)
			self._connections.add(conn_handle)
		elif event == _IRQ_CENTRAL_DISCONNECT:
			conn_handle, _, _ = data
			print("Disconnected", conn_handle)
			self._connections.remove(conn_handle)
            # Start advertising again to allow a new connection.
			self._advertise()
		elif event == _IRQ_GATTS_WRITE:
			# Sample format ssid=name&password=123456789
			conn_handle, value_handle = data
			value = self._ble.gatts_read(value_handle)
			data = value.decode("utf-8")
			self.data += data
			self.write_ssid_password()
	def write_ssid_password(self):
		print(self.data)
		try:
			writeable_data = ujson.loads(self.data)
		except Exception:
			pass
		else:
			# Try to connect to the wifi
			self.data = ""
			if "ssid" not in writeable_data or "password" not in writeable_data:
				return
			try:
				check_connect = do_connect(writeable_data.get("ssid"), writeable_data.get("password"))
			
				if check_connect:
					yes = "CONNECT".encode("utf-8")
					self.send(yes)
					CONFIGURATIONS["ssid"] = writeable_data.get("ssid")
					CONFIGURATIONS["password"] = writeable_data.get("password")
					write_to_config()
					machine.reset()
				else:
					no = "NO CONNECT".encode("utf-8")
					print(no)
					self.send(no)
			except OSError:
				os = "OS ERROR".encode("utf-8")
				print(os)
				self.send(os)
			
	def is_connected(self):
		return len(self._connections) > 0

	def on_write(self, callback):
		self._write_callback = callback
	
	def send(self, data):
		for conn_handle in self._connections:
			self._ble.gatts_notify(conn_handle, self._handle_tx, data)




def main():
	ble = bluetooth.BLE()

	configs = get_configs()

	device_id = configs.get("deviceID")
	print("Device ID: ", device_id)
	assert device_id is not None

	BlePeripheral(ble, name=device_id)


	while True:
		time.sleep_ms(100)

