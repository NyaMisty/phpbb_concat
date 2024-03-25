# phpBB Concat

Concats all pages of a phpBB thread into one

## Feature
- Fully-Auto: Automatically recognize page number
- Concurrent: using 10 threads to concurrently retrieve pages

## Usage

1. Start the program, it will listen at port 3322
2. Access the server via URL: `http://[server_ip]:3322?url=[encoded_url_string]`, for example `http://127.0.0.1:3322?url=https%3A%2F%2Fwww.bogleheads.org%2Fforum%2Fviewtopic.php%3Ft%3D288192%26view%3Dprint`z