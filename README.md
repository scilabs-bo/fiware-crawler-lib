# FIWARE Crawler Library

<!--toc:start-->
- [FIWARE Crawler Library](#fiware-crawler-library)
  - [Installation](#installation)
  - [Usage](#usage)
  - [Configuration](#configuration)
  - [Contributing](#contributing)
  - [License](#license)
<!--toc:end-->

The **fiwarecrawlerlib** is a Go package that provides utility functions for regularly crawling data and creating configuration groups and devices for a FIWARE IoT-Agent UL (Ultralight). It simplifies the integration of IoT devices with FIWARE IoT Agents using the Ultralight protocol for data communication.

## Installation

To use this library in your Go project, you can install it with:

```bash
go get -u github.com/scilabs-bo/fiware-crawler-lib
```

## Usage

```go
package main

import (
    "github.com/scilabs-bo/fiware-crawler-lib"
)

func main() {
    // Create a new instance of the crawler
    crawler := fiwarecrawlerlib.New()

    // Customize the crawler configuration if needed
    // ...

    // Start the crawler job
    crawler.StartJob(func() {
        // Define your crawling logic here
        // ...
    })
}
```

## Configuration

The library relies on the following environment variables for configuration:

| Name            | Required | Default  | Description                                               |
| --------------- | -------- | -------- | --------------------------------------------------------- |
| CRONTAB         | true     | -        | Specifies the cron schedule for data crawling.            |
| IOTA_HOST       | true     | -        | FIWARE IoT-Agent UL host address.                         |
| IOTA_PORT       | true     | -        | FIWARE IoT-Agent UL port number.                          |
| SERVICE         | true     | -        | FIWARE service identifier.                                |
| SERVICE_PATH    | true     | -        | FIWARE service path.                                      |
| API_KEY         | true     | -        | FIWARE API key for authentication.                        |
| RESOURCE        | false    | /iot/d   | Resource path for data storage.                           |
| DEVICE_ID       | true     | -        | Identifier for the IoT device.                            |
| ENTITY_TYPE     | true     | -        | Type of the FIWARE entity associated with the device.     |
| LOG_LEVEL       | false    | DEBUG    | Log level for the library (TRACE, DEBUG, INFO, WARNING, ERROR, FATAL, PANIC). |
| MQTT_BROKER     | false    | mosquitto| MQTT broker address for data publishing.                  |
| MQTT_PORT       | false    | 1883     | MQTT broker port for data publishing.                     |
| CLIENT_ID       | true     | -        | MQTT client identifier.                                   |
| USERNAME        | false    | -        | MQTT broker username (optional).                          |
| PASSWORD        | false    | -        | MQTT broker password (optional).                          |

Adjust the values according to your environment and requirements.

## Contributing

Feel free to contribute to this library by opening issues or pull requests. Your feedback and contributions are highly appreciated.

## License

This project is licensed under the [MIT License](LICENSE).
