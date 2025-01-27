from __future__ import annotations
from abc import ABC, abstractmethod
from typing import Type
import logging

logger = logging.getLogger(__name__)

registered_data_handlers: dict[str, Type[DataHandler]] = {}

class DataHandler(ABC):
    @staticmethod
    def get_handler(name: str) -> Type[DataHandler]:
        logger.info(f'Searching for {name} DataHandler')
        for k,v in registered_data_handlers.items():
            if name in k:
                return v

        raise RuntimeError("Data Handler not found")

    @staticmethod
    def register(name: str, handler: Type[DataHandler]):
        registered_data_handlers[name] = handler

    @abstractmethod
    def set_state(self, state: str) -> str | None:
        pass

    @abstractmethod
    def get_data(self) -> str:
        pass


