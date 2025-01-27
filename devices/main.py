import pika
import proto.messages_pb2 as messages
import proto.services_pb2 as services
import proto.services_pb2_grpc as services_grpc
import actuator
import uuid
from concurrent import futures
from threading import Thread, Lock, Event
from handlers import DataHandler
import pika.connection
import grpc
import sys
import time
import logging
import atexit
import signal

logger = logging.getLogger(__name__)
syslog = logging.StreamHandler()
formatter = logging.Formatter('%(levelname)s %(asctime)s: %(message)s')
syslog.setFormatter(formatter)

logging.basicConfig(handlers=[syslog], level=logging.INFO)
# logger.handlers.clear()
# logger.addHandler(syslog)

broker_connect = Lock()

def connect_to_broker() -> pika.BlockingConnection:
    global broker_connect
    broker_connect.acquire()
    connection = pika.BlockingConnection(pika.ConnectionParameters('localhost'))
    broker_connect.release()
    return connection

class Device():
    def __init__(self, name: str, ip: str, port: str, data: str):
        self.sensor_id: str = str(uuid.uuid4())
        self.actuator_id: str = ""
        self.port = port
        self.ip = ip
        self.name = name
        self.data: DataHandler = DataHandler.get_handler(name)()
        self.data_lock: Lock = Lock()

    def start(self):
        t1 = Thread(target=self._connect_actuator_to_gateway, daemon=False)
        t2 = Thread(target=self.send_to_gateway, daemon=False)
        t1.start()
        t2.start()

    def _connect_actuator_to_gateway(self):
        connection: pika.BlockingConnection = connect_to_broker() 
        channel = connection.channel()
        temp_queue = channel.queue_declare('')

        register_queue = 'connect'
        channel.queue_declare(queue=register_queue)
        registration_order_queue = temp_queue.method.queue

        exchange_name = 'actuator_registration_order_exchange'
        channel.exchange_declare(exchange=exchange_name, exchange_type='fanout')


        channel.queue_bind(exchange=exchange_name, queue=registration_order_queue)
        try:
            while True:
                logger.info("Registering actuator")

                cr = messages.ConnectionRequest(
                    queue_name=f'{self.name}',
                    type=messages.DEVICE_TYPE_ACTUATOR,
                    ip=self.ip,
                    port=self.port,
                    data=self.data.get_data()
                )

                channel.basic_publish(exchange='',
                                      routing_key=register_queue,
                                      body=cr.SerializeToString())

                for method_frame, props, body in channel.consume(registration_order_queue,  auto_ack=True):
                    logger.info("Registration order arrived")
                    break
        except Exception as e:
            logger.error(e)
        finally:
            channel.close()


    def send_to_gateway(self):
        connection: pika.BlockingConnection = connect_to_broker() 
        channel = connection.channel()
        channel.queue_declare(queue='sensor_updates')
        logger.info("Starting sending to gateway")

        try:
            while True:
                sdu = messages.SensorDataUpdate(
                    data=self.data.get_data(),
                    id=self.sensor_id,
                    name=self.name,
                )
                self.data_lock.acquire()
                queue = f'sensor_updates'
                logger.info(f"Sending data to gateway at {queue}")
                logger.info(f"Data: {sdu.data}")
                logger.info(f"ID: {sdu.id}")

                channel.basic_publish(exchange='',
                                      routing_key=queue,
                                      body=sdu.SerializeToString())
                self.data_lock.release()

                time.sleep(5)

        except Exception as err:
            logger.warning(err)
        finally:
            channel.close()

    def change_data(self, request: services.ActuatorState) -> str | None:
        self.data_lock.acquire()
        self.data.set_state(request.state)
        self.data_lock.release()

        return self.data.get_data()

if __name__ == '__main__':
    if len(sys.argv) <= 3:
        print("Usage: python main.py <queue_name> <ip> <port>")
        sys.exit(1)

    queue_name = sys.argv[1]
    ip = sys.argv[2]
    port = sys.argv[3]

    device = Device(queue_name, ip, port, '10')

    device.start()
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    server.add_insecure_port(f"[::]:{port}")
    services_grpc.add_ActuatorServicer_to_server(actuator.ActuatorServer(device), server)
    server.start()
    server.wait_for_termination()
