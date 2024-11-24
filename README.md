# analysis-demo
a demo for Golang Static Analysis

## Background

* lint: It is a demo for Golang AST. It is a simple lint tool to analyze:
  1. Whether there are any identifiers' length is equal to 13.
  2. Whether there are control structures in the code nested more than 4 levels.

* parity: It is a demo for Golang CFG & SSA. It is a simple tool to analyze:
  - The variable is even or odd.

* type check: It is a pluggable type checker for Golang.
  - Use comment to specify the type of the variable.
  - Support custom type checking rules.

* diagnostic: It is a SDK for Golang observation.
  - It can be used to collect the metrics of the code.
  - It can be generated profiling data for the code.
