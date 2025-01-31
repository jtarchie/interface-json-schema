package main

import (
    "encoding/json"
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "os"
    "strings"
)

type Weather interface {
    // Retrieves current weather for the given location
    GetWeather(
        location string, // City and country e.g. Bogotá, Colombia
    ) (temp int)
}

type JSONSchema struct {
    Type     string         `json:"type"`
    Function FunctionSchema `json:"function"`
}

type FunctionSchema struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Parameters  ParameterSchema `json:"parameters"`
    Strict      bool            `json:"strict"`
}

type ParameterSchema struct {
    Type                 string              `json:"type"`
    Properties           map[string]Property `json:"properties"`
    Required             []string            `json:"required"`
    AdditionalProperties bool                `json:"additionalProperties"`
}

type Property struct {
    Type        string   `json:"type"`
    Description string   `json:"description"`
    Enum        []string `json:"enum,omitempty"`
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run main.go <path-to-go-file>")
        return
    }

    filePath := os.Args[1]
    fileSet := token.NewFileSet()
    node, err := parser.ParseFile(fileSet, filePath, nil, parser.ParseComments)
    if err != nil {
        fmt.Println("Error parsing file:", err)
        return
    }

    comments := mapCommentsByLine(node.Comments, fileSet)

    for _, decl := range node.Decls {
        genDecl, ok := decl.(*ast.GenDecl)
        if !ok || genDecl.Tok != token.TYPE {
            continue
        }

        for _, spec := range genDecl.Specs {
            typeSpec, ok := spec.(*ast.TypeSpec)
            if !ok {
                continue
            }

            interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
            if !ok {
                continue
            }

            for _, method := range interfaceType.Methods.List {
                if len(method.Names) == 0 {
                    continue
                }

                methodName := method.Names[0].Name
                funcType, ok := method.Type.(*ast.FuncType)
                if !ok {
                    continue
                }

                parameters := make(map[string]Property)
                required := []string{}
                for _, param := range funcType.Params.List {
                    paramName := param.Names[0].Name
                    paramType := paramTypeToString(param.Type)
                    paramLine := fileSet.Position(param.Pos()).Line
                    paramDescription := comments[paramLine]
                    parameters[paramName] = Property{
                        Type:        paramType,
                        Description: paramDescription,
                    }
                    required = append(required, paramName)
                }

                methodDescription := getCommentText(method.Doc)
                schema := JSONSchema{
                    Type: "function",
                    Function: FunctionSchema{
                        Name:        methodName,
                        Description: methodDescription,
                        Parameters: ParameterSchema{
                            Type:                 "object",
                            Properties:           parameters,
                            Required:             required,
                            AdditionalProperties: false,
                        },
                        Strict: true,
                    },
                }

                jsonData, err := json.MarshalIndent(schema, "", "  ")
                if err != nil {
                    fmt.Println("Error generating JSON:", err)
                    return
                }

                fmt.Println(string(jsonData))
            }
        }
    }
}

func paramTypeToString(expr ast.Expr) string {
    switch t := expr.(type) {
    case *ast.Ident:
        return t.Name
    case *ast.ArrayType:
        return "array"
    case *ast.MapType:
        return "object"
    case *ast.StarExpr:
        return paramTypeToString(t.X)
    default:
        return "unknown"
    }
}

func getCommentText(commentGroup *ast.CommentGroup) string {
    if commentGroup == nil {
        return ""
    }
    var comments []string
    for _, comment := range commentGroup.List {
        comments = append(comments, strings.TrimSpace(strings.TrimPrefix(comment.Text, "//")))
    }
    return strings.Join(comments, " ")
}

func mapCommentsByLine(commentGroups []*ast.CommentGroup, fileSet *token.FileSet) map[int]string {
    comments := make(map[int]string)
    for _, group := range commentGroups {
        for _, comment := range group.List {
            line := fileSet.Position(comment.Slash).Line
            comments[line] = strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
        }
    }
    return comments
}