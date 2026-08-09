# The files these rules are proven against

Every file here is written to be found. The check scans this directory
separately from the tree and requires each rule to report exactly one finding in
it, so a rule that stops biting fails the run instead of going quiet.

Each file also carries the neighbour that must stay silent, one character or one
token away from the finding beside it. A fixture that could not have passed
proves less than one that nearly did, and the mistakes chosen here are the ones
somebody actually makes: a bound copied without its unit, a path pasted out of a
terminal, an address typed without the s.

The tree scan excludes this directory by name. That is a hole the size of one
directory, and it is the price of proving the rules on every run rather than
once in a pull request nobody reads again. Nothing here is read by any other
route and nothing here is a component record.
