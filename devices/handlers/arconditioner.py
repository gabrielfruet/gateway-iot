from typing import Literal
from .data_handler import DataHandler
import random


class ArconditionerDataHandler(DataHandler):
    def __init__(self):
        self.state: Literal["ON", "OFF"] = "OFF"
        self.data_on = 18
        self.data_off = 30

    def set_state(self, state: str) -> str | None:
        state = state.upper()
        match state:
            case "ON":
                self.state = "ON"
                return state
            case "OFF":
                self.state = "OFF"
                return state
            case _:
                return None

    def get_data(self) -> str:
        match self.state:
            case "ON":
                return str(self.data_on + random.uniform(-2,2))
            case "OFF":
                return str(self.data_off + random.uniform(-2,2))


DataHandler.register('arconditioner', ArconditionerDataHandler)
