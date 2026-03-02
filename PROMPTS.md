# Future prompts to line up for agent work

* When task board is json type. we must save that taskboard to postgres. worker must pull from postgres. we need to reimplement the local json filesystem taskboard now and move it fully into postgres for distrubuted sharing. We are also going to rename it from json to internal. ensure no part of the now named "internal" tracker is on filesystem. Remove all filesystem code related to the old json tracker.

* Change priority from string to int to make priority filtering easier higher the number the higher the priority. Convert string priorities from ingestion to int.

* Drop all implemenations of Tracker Ingestion implementations like our github issues implementations. or any others present in v1 or desktop client. so that were left with creating the project and board and not be able ingest anything to it yet. this will include dropping all env config for api and worker related to tracker boards it will now solely be configured by the user at project creation time.
