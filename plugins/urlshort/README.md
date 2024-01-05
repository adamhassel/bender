# URL shortener

This plugin will grab any URL posted in a channel the bot is in and ask bitly, tinyurl or cleanuri for a short version. The minimum length of the URL the bot should shorten is configurable.
You will need an API token for bitly or tinyurl as well, which you need to put in the config file. If "service" is undefined, default is cleanuri.
If you're not using an api-key, set it to a non-empty dummy value.

As an added bonus, (can  be disabled with "cleanup=false" in config) the URL sent to the shortening service is stripped of a wide range of tracking parameters, which are defined in `tracking.json`.