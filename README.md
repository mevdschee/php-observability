# php-observability

To run the server:

    go run .

In bash run:

    for run in {1..100}; do php writer.php & done

And to stop:

    killall php

Now observe the stats:

http://localhost:4000/

Enjoy!