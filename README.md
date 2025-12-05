# Geometric Constraint Solver
The only toy geometric constraint solver on the internet. Simplified for easy understanding and using only the Go standard library, so there is no magic, just about 2000 lines of code. You can step through it and understand how CAD programs work. 

It is inspired by [SolveSpace](https://solvespace.com/), but built independently from the ground up. Geometric constraints are represented as a system of nonlinear equations and solved by multidimensional [Newton's method](https://en.wikipedia.org/wiki/Newton%27s_method). I built a simple [symbolic algebra system](https://en.wikipedia.org/wiki/Computer_algebra_system) that calculates derivatives by traversing the syntax tree. To keep the codebase simple, this implementation assumes that the system will converge.

I included a short [math summary](docs/math.md).

## Demo
https://github.com/user-attachments/assets/eb8d087f-5ca4-4fac-9ae4-b68d3771c1d4

## Running the program

`go run ./cmd`
