# Future prompts to line up for agent work

* When task board is json type. we must save that taskboard to postgres. worker must pull from postgres. we need to reimplement the local json filesystem taskboard now and move it fully into postgres for distrubuted sharing. We are also going to rename it from json to internal. ensure no part of the now named "internal" tracker is on filesystem. Remove all filesystem code related to the old json tracker.
