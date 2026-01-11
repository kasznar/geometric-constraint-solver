# Overview of The Solver

## Project Structure
The Solver built up form four packages:
* `math`: Basic Vector and Matrix arithmetics. 
* `solver`:
    * `expr`: Abstract Syntax Tree to represent equations and tree traversal algorithms for evaluation and partial differentiation.
    * `solver`: Equation system representation and solving algorithm. 
* `sketch`:
* `utils`: Utility functions for testing. 

Additionally the `cmd` package has a minimal UI for testing out the solver.