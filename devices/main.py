import pika
import proto.messages_pb2 as messages
import proto.services_pb2 as services
import proto.services_pb2_grpc as services_grpc
import actuator
from concurrent import futures
from threading import Thread, Lock
import grpc
import sys
import time
import logging
import atexit
import signal

logger = logging.getLogger(__name__)
syslog = logging.StreamHandler()
formatter = logging.Formatter('%(levelname)s %(asctime)s %(app_name)s: %(message)s')
syslog.setFormatter(formatter)
logger.setLevel(logging.INFO)
logger.handlers.clear()
logger.addHandler(syslog)


def connect_to_broker() -> pika.BlockingConnection:
    connection = pika.BlockingConnection(pika.ConnectionParameters('localhost'))
    return connection

class Device():
    def __init__(self, name: str, data: int):
        self.connection: pika.BlockingConnection = connect_to_broker() 
        self.id: str | None = None
        self.name = name
        self.data: int = data
        self.data_lock: Lock = Lock()

    def start(self):
        self._connect_to_gateway()
        t = Thread(target=self.send_to_gateway)
        t.run()

    def _connect_to_gateway(self):
        channel = self.connection.channel()

        channel.queue_declare(queue='connect')
        channel.queue_declare(queue=self.name)
        channel.queue_declare(queue=f'{self.name}_id')

        cr = messages.ConnectionRequest(queue_name=self.name, type=messages.DEVICE_TYPE_SENSOR)

        channel.basic_publish(exchange='',
                              routing_key='connect',
                              body=cr.SerializeToString())


        for method_frame, props, body in channel.consume(f'{self.name}_id',  auto_ack=True):
            self.id = messages.ConnectionResponse.FromString(body).id
            break

        channel.close()

    def send_to_gateway(self):
        channel = self.connection.channel()

        try:
            while True:
                sdu = messages.SensorDataUpdate(data=str(self.data),id=self.id)
                self.data_lock.acquire()
                logger.info("Sending data to gateway")
                logger.info(f"Data: {sdu.data}")
                logger.info(f"ID: {sdu.id}")

                channel.basic_publish(exchange='',
                                      routing_key=queue_name,
                                      body=sdu.SerializeToString())
                self.data_lock.release()

                time.sleep(5)

        except Exception as err:
            logger.warning(err)
        finally:
            channel.close()

    def change_data(self, request: services.ActuatorState) -> int | None:
        if request.id != self.id:
            logger.warning("ID does not match")
            return None

        self.data_lock.acquire()
        self.data = int(request.state)
        self.data_lock.release()

        return self.data


    def disconnect(self):
        channel = self.connection.channel()

        channel.queue_declare("disconnect")

        dr = messages.DisconnectionRequest(id=self.id, queue_name=self.name)

        logger.info("Sending disconnect message")

        channel.basic_publish(
            "",
            "disconnect",
            dr.SerializeToString()
        )

        channel.close()
        self.connection.close()



def handle_signal(signum, frame):
    print(f"Signal {signum} received, exiting gracefully.")
    sys.exit(0)


if __name__ == '__main__':
    signals = [signal.SIGINT, signal.SIGTERM, signal.SIGHUP]
    for sig in signals:
        signal.signal(sig, handle_signal)

    if len(sys.argv) <= 1:
        print("Usage: python main.py <queue_name>")
        sys.exit(1)

    queue_name = sys.argv[1]

    logger = logging.LoggerAdapter(logger, {"app_name":queue_name})

    device = Device(queue_name, 10)

    atexit.register(device.disconnect)

    try:
        device.start()
        server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
        server.add_insecure_port("[::]:50051")
        services_grpc.add_ActuatorServicer_to_server(actuator.ActuatorServer(device), server)
        server.start()
        server.wait_for_termination()
    except Exception:
        pass
    finally:
        sys.exit(0)
