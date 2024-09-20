# php-observability

To run the server:

    go run .

In bash run:

    for run in {1..100}; do php writer.php & done

And to stop:

    killall php

Now observe the stats:

http://localhost:4000/

Example stats:

    database_calls_count{file="file.php@L123"} 1000
    database_calls_seconds{file="file.php@L123"} 142.857

Enjoy!