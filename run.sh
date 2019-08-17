#! /bin/bash

from=${P2O_FROM?"No start date provided"}
to=${P2O_TO?"No end date provided"}

./exporter -e /etc/p2o/exporter.json -q /etc/p2o/query.json -f $from -t $to