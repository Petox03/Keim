package main

import (
	"fmt"
	"os"
	/* "path/filepath" */

	"keim/internal/project"
	"keim/internal/validator"
	"keim/internal/templates"
	"keim/internal/generator"
)

func main() {

	projectData := project.Project{
		Name: 		"Templating",
		Path: 		"./test/",
		GoVersion: 	"1.26",
	}

	files := templates.FileNames()

	fmt.Println("--- Probando mi validador en vivo ---")

	if err := validator.Validate(projectData.Path, files); err != nil {
		fmt.Printf("[Validator]: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("La carpeta está limpia.")

	fmt.Println("--- Iniciando prueba de templates embebidas ---")
	fmt.Println("Arcvhivos a generar: ")
	for _, file := range files {
		fmt.Printf("%v\n", file)
	}

	err := generator.Generate(projectData, files)
	if err != nil {
		fmt.Printf("[Generator]: %v", err)
	}
	fmt.Printf("El projecto %v ha sido creádo en la carpeta %v.", projectData.Name, projectData.Path)

}