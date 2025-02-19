import random
from .data_handler import DataHandler
from typing import Literal
import csv

class LightDataHandler(DataHandler):
    def __init__(self):
        self.state: Literal["ON", "OFF"] = "OFF"
        self.brightness = 0  # 0 to 100

    def set_state(self, state: str) -> str | None:
        state = state.upper()
        if "|" in state:
            parts = state.split("|")
            if len(parts) == 2 and parts[0] in {"ON", "OFF"}:
                try:
                    brightness = int(parts[1])
                    if 0 <= brightness <= 100:
                        state = parts[0]

                        if state in ("ON", "OFF"):
                            self.state = state

                        self.brightness = brightness if self.state == "ON" else 0
                        return state
                except ValueError:
                    return None
        elif state in ("ON", "OFF"):
            self.state = state
            if self.state == "OFF":
                self.brightness = 0
            return state

        return None

    def get_data(self) -> str:
        return str(self.brightness) if self.state == "ON" else "0"


class TemperatureSensorDataHandler(DataHandler):
    def __init__(self):
        self.base_temp = 22  # Default room temperature

    def set_state(self, state: str) -> str | None:
        return None  # Temperature sensor has no state to set

    def get_data(self) -> str:
        return str(self.base_temp + random.uniform(-1.5, 1.5))


class DoorLockDataHandler(DataHandler):
    def __init__(self):
        self.state: Literal["LOCKED", "UNLOCKED"] = "LOCKED"

    def set_state(self, state: str) -> str | None:
        state = state.upper()
        if state in ("LOCKED", "UNLOCKED"):
            self.state = state
            return state
        return None

    def get_data(self) -> str:
        return self.state

class CarLocalizationDataHandler(DataHandler):
    def __init__(self):
        self.coordinates = []
        self.step = 1
        self.index = 0
        with open('handlers/data/coordinates.csv', mode="r") as file:
            reader = csv.reader(file)
            next(reader)  # Skip the header
            for row in reader:
                x, y = map(float, row)
                self.coordinates.append((x, y))

        for i in range(len(self.coordinates)-1,-1,-1):
            self.coordinates.append(self.coordinates[i])

    def set_state(self, state: str) -> str | None:
        return None  # CarLoc sensor has no state to set

    def get_data(self) -> str:
        self.index+=self.step
        data = self.coordinates[self.index%len(self.coordinates)]
        return f"{data[0]}|{data[1]}"


DataHandler.register('light', LightDataHandler)
DataHandler.register('temperature_sensor', TemperatureSensorDataHandler)
DataHandler.register('door_lock', DoorLockDataHandler)
DataHandler.register('car_loc', CarLocalizationDataHandler)

