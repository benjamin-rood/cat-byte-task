### Application 1: API (please use GIN library)

ADD POST Endpoint: /message

With POST body: `{ sender: String, receiver: String, message: String }`

-> pushes received information to a RabbitMQ Queue Return OK Status if everything is there, otherwise Bad Request

### Application 2: MessageProcessor

Subscribes to queue from RabbitMQ and processes the message

Processing of message means saving the message to Redis in a way such that application 3 will work.

### Application 3: Reporting API

ADD GET Endpoint: /message/list

With Query Parameters: `sender: String, receiver: String`

Returns an array of objects with the following fields:
- sender
- receiver
- message (content that was exchanged between sender and receiver in chronological descending order)