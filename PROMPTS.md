# Future prompts to line up for agent work

* When task board is json type. we must save that taskboard to postgres. worker must pull from postgres. we need to reimplement the local json filesystem taskboard now and move it fully into postgres for distrubuted sharing. We are also going to rename it from json to internal. ensure no part of the now named "internal" tracker is on filesystem. Remove all filesystem code related to the old json tracker.

* Change priority from string to int to make priority filtering easier higher the number the higher the priority. Convert string priorities from ingestion to int.

* Drop all implemenations of Tracker Ingestion implementations like our github issues implementations. or any others present in v1 or desktop client. so that were left with creating the project and board and not be able ingest anything to it yet. this will include dropping all env config for api and worker related to tracker boards it will now solely be configured by the user at project creation time.

* The scm provider is not something configured by api/worker at boot time it's during client project creation/management. which means securly storing the PAT we require & creating an scm provider repository to store and reference them in the project. this will currently be a per project scope of configuring scm provider & credentials. Ensure you touch all code in v1 worker/api & the frontend desktop. ensure you clean all code that does not comply with this leaving no track of how it's currently bootstrapped
