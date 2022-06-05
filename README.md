# kubectl-common

A simple kubectl plugin for creating commands for common tasks.

## Usage

Add the command:  
`kubectl common add --name ls-prom --command="kubectl get pods -A | grep prometheus"`

Use it:  
`kubectl common ls-prom`  
Lists the pods that contain prometheus in their names

You don't need to use it for kubernetes commands. It can be any command you want but that defeats the purpose. Usually the long commands that I use comes from kubernetes and making it a plugin like that feels natural to use.
