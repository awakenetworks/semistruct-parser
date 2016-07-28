# Welcome!

`semistruct-parse` is a project providing a parser for semistructured lines of
text. These are primarily used in a custom journald logging
driver for the purpose of feeding semistructured data into journald from docker.

## The Format

Here is an example of a semistructured line that, when parsed by the top-level
`ParseSemistruct` combinator, emits a struct of standard Go data structures
representing the three primary fields:

`!< 8 [db3:slave] { region="us-east-d zone=2" uuid=3c79bfc0-4b27-4c99-a952-8f435bbf0514 } >!`

The parser looks for the beginning sentinel character sequence `!<` and finishes
when it finds the ending sentinel character sequence `>!`.
