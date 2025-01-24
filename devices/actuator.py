import main
import grpc
import proto.services_pb2_grpc as services_grpc
import proto.services_pb2 as services

class ActuatorServer(services_grpc.ActuatorServicer):
    def __init__(self, device: main.Device):
        self.device=device


    def ChangeState(self, request: services.ActuatorState, context):
        data = self.device.change_data(request)
        if data is None:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details("ID does not match")
            return services.ActuatorState(id=self.device.id, state="0")

        services.ActuatorState(id=self.device.id, state=str(data))
