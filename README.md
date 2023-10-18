# vcpot-demo
Demo code for vector commitment based proof of transit

## Construction
The name "vector commitment" is a broader term that refers to a commitment scheme for a vector of values. It has many different constructions, from trivial ones like merkle tree to more complex ones like kzg polynomial commitment and verkle tree. The difference lies in the efficiency to commit the vector (compute commitment), open a specific value (compute single opening proof), verify opening proof against the commitment. 

In this repo/solution, we use kzg polynomial commitment. For details, I refer you to Vitalik's blog on [Verkle Trees](https://vitalik.ca/general/2021/06/18/verkle.html) and Dankrad's blog on [kzg polynomial commitments](https://dankradfeist.de/ethereum/2020/06/16/kate-polynomial-commitments.html). 


## Usage

- To run the demo, do: 

    `bash run.sh` and follow the instructions.

- To terminate the demo, do: 

    `bash kill.sh` in another terminal.

## Version Requirements
Recommend to use brew for mac
- bash > 4.x
- go > 1.21
- tmux > 3.3
