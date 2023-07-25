package main

import "text2phenotype.com/fdl/logger"

func main() {
	logger.WrapProcess("go", "run", "./entrypoint")
}
