# Gateway IOT

## Setup

To setup this project, execute this:

```bash
make setup
```

## Usage

### Running RabbitMQ

```bash
make rabbitmq
```
### Running the Gateway

At the root of the project, do:

```bash
cd gateway; go run .
```

### Running the Device

```bash
python devices/main.py <device_name> <ip> <grpc_port>
```

The `device_name` has to have a DataHandler registered for it.

Here’s a concise REST API description for the provided code in Markdown format:

# REST API Description

## 1. Get Actuators
- **Route**: `/actuators`
- **Verb**: `GET`
- **Description**: Retrieves a list of all actuators.
- **Response**: JSON array of actuator names.

## 2. Change Actuator State
- **Route**: `/actuators`
- **Verb**: `POST`
- **Description**: Changes the state of a specific actuator.
- **Request Body**:
  ```json
  {
    "name": "string", // Actuator name
    "state": "string" // New state
  }
  ```
- **Response**: No content.

## 3. Get Sensors
- **Route**: `/sensors`
- **Verb**: `GET`
- **Description**: Retrieves a list of all sensors.
- **Response**: JSON array of sensor names.

## 4. Get Sensor Data
- **Route**: `/sensors?name=<sensor_name>`
- **Verb**: `GET`
- **Description**: Retrieves data for a specific sensor.
- **Query Parameter**:
  - `name`: Name of the sensor.
- **Response**: JSON object containing sensor data.

## Notes
- All responses are in JSON format.
- Errors return appropriate HTTP status codes with error messages.
