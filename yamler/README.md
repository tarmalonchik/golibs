## Use this tool to generate `.golangci.yml` using template

## Installation
`go install gitlab.diftech.org/pos/go-possum/yamler`

## Usage
`yamler --o .golangci.yml --е .golangci-tmpl.yml`

## Help

`yamler --help`

## Available yaml options

- ### To append data to an existing array, ensure that the first element in the new array is <br>
```yml
- (( append ))
```
<ul>
Example:

`base.yml`
```
linters:
  default: none
  enable:
    - asciicheck
    - bidichk
    - bodyclose
```
`[template file] .golangci-tmpl.yml`
```
linters:
  enable:
    - (( append ))
    - somenewlinter
```
`[expected output] .golangci.yml`
```
linters:
  default: none
  enable:
  - asciicheck
  - bidichk
  - bodyclose
  - somenewlinter
```
</ul>

- ### To delete data from an existing array, ensure that the first element in the new array is <br>
```yml
- (( delete [element-to-delete]))
```
<ul>
Example:

`base.yml`
```
linters:
  default: none
  enable:
    - asciicheck
    - bidichk
    - bodyclose
    - pgclose
```
`[template file] .golangci-tmpl.yml`
```
linters:
  enable:
    - (( delete asciicheck ))
    - (( delete pgclose ))
```
`[expected output] .golangci.yml`
```
linters:
  default: none
  enable:
  - bidichk
  - bodyclose
```
</ul>

- ### To delete key, use the following way <br>
<ul>
Example:

`base.yml`
```
linters:
  default: none
run:
  timeout: 10m
  relative-path-mode: gomod
  concurrency: 4
```
`[template file] .golangci-tmpl.yml`
```
linters:
  default: # this will be interpreted as empty value
```
`[expected output] .golangci.yml`
```
linters:
  default: null
run:
  concurrency: 4
  relative-path-mode: gomod
  timeout: 10m

```
</ul>

- ### To extend keys, use the following way <br>
<ul>
Example:

`base.yml`
```
run:
  timeout: 10m
  relative-path-mode: gomod
  concurrency: 4
```
`[template file] .golangci-tmpl.yml`
```
run:
  new_key_will_be_added: new_value
```
`[expected output] .golangci.yml`
```
run:
  concurrency: 4
  new_key_will_be_added: new_value
  relative-path-mode: gomod
  timeout: 10m


```
</ul>

## Docs
To see more options, use the following link 
https://github.com/geofffranks/spruce/tree/main/doc