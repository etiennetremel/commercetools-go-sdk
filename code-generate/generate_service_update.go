package main

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

// Generate the `<service>Update` function
func generateServiceUpdate(funcName string, resourceName string, resourceService ResourceService, resourceMethod ResourceMethod, httpMethod ResourceHTTPMethod) (code *jen.Statement) {
	resourceIdentifier := createResourceIdentifier(resourceService, resourceMethod)
	updateStructID := fmt.Sprintf("%sInput", funcName)
	updateObjectType := fmt.Sprintf("%sUpdateAction", resourceName)

	// Generate the `type <resource>With<ID>Input struct {}`
	c := jen.Commentf("%s is input for function %s", updateStructID, funcName).Line()
	c.Type().Id(updateStructID).Struct(
		jen.Id(resourceIdentifier.PublicName).String(),
		jen.Id("Version").Int(),
		jen.Id("Actions").Index().Id(updateObjectType),
	).Line()

	c.Func().Params(jen.Id("input").Op("*").Id(updateStructID)).Id("Validate").Params().Parens(jen.Id("error")).Block(

		// Validate if input identifier is empty
		jen.If(jen.Id("input").Op(".").Id(resourceIdentifier.PublicName).Op("==").Lit("")).Block(
			jen.Return(
				jen.Qual("fmt", "Errorf").Call(
					jen.Lit(fmt.Sprintf("no valid value for %s given", resourceIdentifier.PublicName)),
				),
			),
		),

		// Validate if update actions are empty
		jen.If(jen.Id("len").Call(jen.Id("input").Op(".").Id("Actions")).Op("==").Lit(0)).Block(
			jen.Return(
				jen.Qual("fmt", "Errorf").Call(
					jen.Lit("no update actions specified"),
				),
			),
		),

		jen.Return(jen.Nil()),
	).Line()

	methodParams := jen.List(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("input").Op("*").Id(updateStructID),
		jen.Id("opts").Op("...").Id("RequestOption"),
	)
	clientMethod := "Update"

	structReceiver := jen.Id("client").Op("*").Id("Client")
	description := fmt.Sprintf("for type %s", resourceService.ResourceType)
	if httpMethod.Description != "" {
		description = httpMethod.Description
	}
	returnParams := jen.List(
		jen.Id("result").Op("*").Id(resourceService.ResourceType),
		jen.Err().Error(),
	)

	c.Commentf("%s %s", funcName, description).Line()
	c.Func().Params(structReceiver).Id(funcName).Params(methodParams).Parens(returnParams).Block(
		// if err := input.Validate(); err != nil {
		// 	return nil, err
		// }
		jen.If(
			jen.Err().Op(":=").Id("input").Op(".").Id("Validate").Call(),
			jen.Err().Op("!=").Nil(),
		).Block(
			jen.Return(jen.Nil(), jen.Err()),
		).Line(),

		// params := url.Values{}
		// for _, opt := range opts {
		// 	opt(&params)
		// }
		jen.Id("params").Op(":=").Qual("net/url", "Values").Block(),
		jen.For(jen.List(jen.Id("_"), jen.Id("opt")).Op(":=").Range().Id("opts")).Block(
			jen.Id("opt").Call(jen.Op("&").Id("params")),
		).Line(),

		resourceIdentifier.createEndpointCode(true),
		jen.Id("err").Op("=").Id("client").Op(".").Id(clientMethod).Call(
			jen.Id("ctx"),
			jen.Id("endpoint"),
			jen.Id("params"),
			jen.Id("input").Op(".").Id("Version"),
			jen.Id("input").Op(".").Id("Actions"),
			jen.Op("&").Id("result"),
		),
		jen.If(jen.Err().Op("!=").Nil()).Block(
			jen.Return(jen.Nil(), jen.Err()),
		),
		jen.Return(jen.Id("result"), jen.Nil()),
	).Line()

	return c
}
