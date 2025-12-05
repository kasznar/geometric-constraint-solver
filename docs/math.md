# Math Summary
## Gauss Elimination
Given a system of linear equations:  
$A\vec{x} = \vec{b}$

Gauss elimination transforms the system into an upper triangular form using row operations.

Steps:
1. Forward elimination:  
   Eliminate variables from lower rows to form an upper triangular matrix $U$.
2. Back substitution:  
   Solve for $\vec{x}$ starting from the last equation upwards.

The result is a solution $\vec{x}$ such that $A\vec{x} = \vec{b}$, if the system is consistent and $A$ is non-singular.

## Least Squares
Given an overdetermined system:  
$A\vec{x} \approx \vec{b}$

The least squares solution minimizes the squared error:  
$\min_{\vec{x}} \| A\vec{x} - \vec{b} \|^2$

The normal equation is:  
$A^T A \vec{x} = A^T \vec{b}$

Solving for $\vec{x}$:  
$\vec{x} = (A^T A)^{-1} A^T \vec{b}$


## Newton's Method

Newton's method can be extended to solve systems of $n$ nonlinear equations in $n$ variables:

$$
\mathbf{F}(\mathbf{x}) = 
\begin{bmatrix}
f_1(x_1, x_2, \dots, x_n) \\
f_2(x_1, x_2, \dots, x_n) \\
\vdots \\
f_n(x_1, x_2, \dots, x_n)
\end{bmatrix}
= \mathbf{0}
$$

The goal is to find $\mathbf{x} = \begin{bmatrix} x_1 & x_2 & \dots & x_n \end{bmatrix}^T$ such that $\mathbf{F}(\mathbf{x}) = 0$.

## Iterative Formula

Given an initial guess $\mathbf{x}^{(0)}$, iterate:

$$
\mathbf{x}^{(k+1)} = \mathbf{x}^{(k)} - J^{-1}(\mathbf{x}^{(k)}) \mathbf{F}(\mathbf{x}^{(k)})
$$

where $J(\mathbf{x})$ is the Jacobian matrix:

$$
J(\mathbf{x}) = \begin{bmatrix}
\frac{\partial f_1}{\partial x_1} & \cdots & \frac{\partial f_1}{\partial x_n} \\
\vdots & \ddots & \vdots \\
\frac{\partial f_n}{\partial x_1} & \cdots & \frac{\partial f_n}{\partial x_n}
\end{bmatrix}
$$

## Need to solve
Solve $ J(\mathbf{x}^{(k)}) \mathbf{d}^{(k)} = \mathbf{F}(\mathbf{x}^{(k)}) $ for $ \mathbf{d}^{(k)} $

### Steps

1. Choose an initial guess $ \mathbf{x}^{(0)} $
2. Evaluate $ \mathbf{F}(\mathbf{x}^{(k)}) $
3. Compute the Jacobian $ J(\mathbf{x}^{(k)}) $
4. Solve $ J(\mathbf{x}^{(k)}) \mathbf{d}^{(k)} = \mathbf{F}(\mathbf{x}^{(k)}) $ for $ \mathbf{d}^{(k)} $
5. Update $ \mathbf{x}^{(k+1)} = \mathbf{x}^{(k)} - \mathbf{d}^{(k)} $
6. Repeat until convergence