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

```bash
make gateway
```

### Running the Device

```bash
make devices
```

On `devices/launch.sh`, you can configure the devices that will be launched

```bash
...

devices=(
    "arconditioner-1" 
    "arconditioner-2" 
    "arconditioner-3" 
    "light-1" 
    "temperature_sensor-2" 
    "door_lock-3"
    ... add more data handlers/devices
)

...
```


# REST API Description

## Get Actuators
- **Route**: `/actuators`
- **Verb**: `GET`
- **Description**: Retrieves a list of all actuators.
- **Response**: JSON array of actuator names.

## Get Actuators State
- **Route**: `/actuators?name=<actuator_name>`
- **Verb**: `GET`
- **Description**: Retrieves the actuator state.
- **Response**: String of actuator state.

## Change Actuator State
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

## Get Sensors
- **Route**: `/sensors`
- **Verb**: `GET`
- **Description**: Retrieves a list of all sensors.
- **Response**: JSON array of sensor names.

## Get Sensor Data
- **Route**: `/sensors?name=<sensor_name>`
- **Verb**: `GET`
- **Description**: Retrieves data for a specific sensor.
- **Query Parameter**:
  - `name`: Name of the sensor.
- **Response**: JSON object containing sensor data.
