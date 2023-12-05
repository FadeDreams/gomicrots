## Golang Microservices Application

This Golang application introduces a robust microservices architecture with key components for efficient user authentication, comprehensive logging, and seamless communication.

## Overview

### Auth Microservice
Efficiently manages user authentication using JWT, ensuring secure access to the system. This Golang-based microservice prioritizes user security and access control.

### Loghandler Microservice
Plays a pivotal role in logging activities across other microservices, promoting effective monitoring and troubleshooting. This component enhances the system's observability and aids in identifying and resolving issues.

### Communication via RabbitMQ
These microservices communicate seamlessly through RabbitMQ, establishing a scalable and decoupled architecture. RabbitMQ facilitates reliable and asynchronous communication between the different components of the system.

## Project Structure
```plaintext
/gomicrots
|-- amqp1
|-- auth
|-- client
|-- docker-compose.yml
|-- loghandler
|-- main
|-- react_client
|-- service1
```

## Getting Started
1. Clone the repository.
2. Run the Docker Compose configuration using `docker-compose up`.
3. Explore the various microservices and their functionalities.

Feel free to contribute, report issues, or provide feedback to enhance and optimize this Golang microservices application. Let's build a secure, well-structured, and efficient system together!
