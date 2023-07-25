# How to start


* Install Go v1.15
* `git clone git@github.com:text2phenotype/ctakes-go.git`
* `dvc pull`
* `cd src`
* `go get && go run ./.`

YAML configurations folder - `./config`

Dictionaries folder - `./resources/dictionaries`

Environment example:

```yaml
MDL_COMN_RMQ_HOST=0.0.0.0
MDL_COMN_RMQ_PORT=5672
MDL_COMN_RMQ_USERNAME=rmquser
MDL_COMN_RMQ_PASSWORD=rmquser
MDL_COMN_RMQ_DEFAULT_EXCHANGE=text2phenotype-default-exchange

MDL_COMN_FDL_TASK_QUEUE=task-fdl
MDL_COMN_SEQUENCER_TASK_QUEUE=task-sequencer

MDL_COMN_STORAGE_CONTAINER_NAME=
MDL_COMN_AWS_ACCESS_ID=
MDL_COMN_AWS_ACCESS_KEY=
MDL_COMN_AWS_REGION_NAME=

# Possible values: DEBUG, INFO, WARN, ERROR, FATAL, PANIC
MDL_COMN_LOGLEVEL=INFO

# Path to dictionaries folder (the same as for cTakes)
FDL_DICTIONARY_PATH=/path_to_repo/resources/dictionaries
FDL_MQ_MAX_PARALLEL_REQUESTS=6
# YAML configurations folder path
FDL_CONFIG_PATH=/path_to/go/config

# Path to this project folder
FDL_DIR_PATH=.

#REST API
FDL_REST_API_ACTIVE=false
FDL_REST_API_PORT=10000
```

### Config example ###
```YAML
params:
  # parameters for the FDL component
  FDL:
    term_dictionary: snomedct/snomedct_term.bsv
    CUI_IDX: 0
    TERM_IDX: 1
    concept_dictionary: snomedct/snomedct_concept.bsv
    scheme: CUI|TUI|CODE|SAB|PREF
    precision_mode: True

# pipeline type (allowed: default_clinical, smoking_status)
pipeline: default_clinical

features:
  - lab         # lab attributes are required
  - drug        # drug attributes are required
  - polarity#   # polarity attributes are required
```
