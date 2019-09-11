// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag at
// 2019-09-11 12:43:35.000071 +0800 CST m=+0.927325273

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "mylxsw",
            "url": "https://github.com/mylxsw/sync",
            "email": "mylxsw@aicode.cc"
        },
        "license": {
            "name": "MIT",
            "url": "https://raw.githubusercontent.com/mylxsw/sync/master/LICENSE"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/": {
            "get": {
                "summary": "欢迎页面，API版本信息",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/controller.WelcomeMessage"
                        }
                    }
                }
            }
        },
        "/failed-jobs/": {
            "get": {
                "tags": [
                    "FailedJobs"
                ],
                "summary": "返回失败的所有任务",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/job.FileSyncJob"
                            }
                        }
                    }
                }
            }
        },
        "/failed-jobs/{id}/": {
            "put": {
                "tags": [
                    "FailedJobs"
                ],
                "summary": "重试失败的任务",
                "parameters": [
                    {
                        "type": "string",
                        "description": "要重试的 Job ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/job.FileSyncJob"
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "FailedJobs"
                ],
                "summary": "删除失败的任务",
                "parameters": [
                    {
                        "type": "string",
                        "description": "删除失败的 Job ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/job.FileSyncJob"
                        }
                    }
                }
            }
        },
        "/histories/": {
            "get": {
                "tags": [
                    "Histories"
                ],
                "summary": "查询最近的文件同步记录",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "返回记录数目",
                        "name": "limit",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/controller.History"
                            }
                        }
                    }
                }
            }
        },
        "/histories/{id}/": {
            "get": {
                "tags": [
                    "Histories"
                ],
                "summary": "返回指定ID的历史纪录详情",
                "parameters": [
                    {
                        "type": "string",
                        "description": "记录ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/controller.History"
                        }
                    }
                }
            }
        },
        "/jobs-bulk/": {
            "post": {
                "tags": [
                    "Jobs"
                ],
                "summary": "批量发起文件同步",
                "parameters": [
                    {
                        "description": "同步定义列表",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "$ref": "#/definitions/controller.BulkSyncReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/controller.JobStatus"
                            }
                        }
                    }
                }
            }
        },
        "/jobs/": {
            "get": {
                "tags": [
                    "Jobs"
                ],
                "summary": "返回队列中所有任务",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/job.FileSyncJob"
                            }
                        }
                    }
                }
            },
            "post": {
                "tags": [
                    "Jobs"
                ],
                "summary": "发起文件同步任务",
                "parameters": [
                    {
                        "type": "string",
                        "description": "同步定义名称",
                        "name": "def",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/controller.JobStatus"
                        }
                    }
                }
            }
        },
        "/jobs/{id}/": {
            "get": {
                "tags": [
                    "Jobs"
                ],
                "summary": "查询文件同步任务执行状态",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Job ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/controller.JobStatus"
                        }
                    }
                }
            }
        },
        "/running-jobs/{id}/": {
            "get": {
                "tags": [
                    "RunningJobs"
                ],
                "summary": "运行中的任务状态，基于websocket"
            }
        },
        "/setting/global-sync/": {
            "get": {
                "tags": [
                    "Setting"
                ],
                "summary": "全局同步配置",
                "parameters": [
                    {
                        "type": "string",
                        "description": "输出格式：yaml/json",
                        "name": "format",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/meta.GlobalFileSyncSetting"
                        }
                    }
                }
            },
            "post": {
                "tags": [
                    "Setting"
                ],
                "summary": "更新全局同步配置",
                "parameters": [
                    {
                        "description": "全局同步定义",
                        "name": "def",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "$ref": "#/definitions/meta.GlobalFileSyncSetting"
                        }
                    }
                ]
            }
        },
        "/sync-bulk/": {
            "delete": {
                "tags": [
                    "Sync"
                ],
                "summary": "批量删除同步定义",
                "parameters": [
                    {
                        "description": "定义名称列表",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "$ref": "#/definitions/controller.BulkDeleteReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/sync/": {
            "get": {
                "tags": [
                    "Sync"
                ],
                "summary": "查询所有文件同步定义",
                "parameters": [
                    {
                        "type": "string",
                        "description": "输出格式：yaml/json",
                        "name": "format",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/meta.FileSyncGroup"
                            }
                        }
                    }
                }
            },
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Sync"
                ],
                "summary": "更新文件同步定义",
                "parameters": [
                    {
                        "description": "文件同步定义",
                        "name": "def",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/meta.FileSyncGroup"
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/meta.FileSyncGroup"
                            }
                        }
                    }
                }
            }
        },
        "/sync/{name}/": {
            "get": {
                "tags": [
                    "Sync"
                ],
                "summary": "查询单个文件同步定义",
                "parameters": [
                    {
                        "type": "string",
                        "description": "定义名称",
                        "name": "name",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "输出格式：yaml/json",
                        "name": "format",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/meta.FileSyncGroup"
                            }
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "Sync"
                ],
                "summary": "删除单个文件同步定义",
                "parameters": [
                    {
                        "type": "string",
                        "description": "定义名称",
                        "name": "name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "action.Factory": {
            "type": "object"
        },
        "collector.Collector": {
            "type": "object",
            "properties": {
                "collectors": {
                    "type": "object",
                    "$ref": "#/definitions/collector.Collectors"
                },
                "index": {
                    "type": "integer"
                },
                "jobID": {
                    "type": "string"
                },
                "lock": {
                    "type": "string"
                },
                "stages": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/collector.Stage"
                    }
                }
            }
        },
        "collector.Collectors": {
            "type": "object",
            "properties": {
                "collectors": {
                    "type": "object"
                },
                "lock": {
                    "type": "string"
                }
            }
        },
        "collector.Progress": {
            "type": "object",
            "properties": {
                "lock": {
                    "type": "string"
                },
                "max": {
                    "type": "integer"
                },
                "total": {
                    "type": "integer"
                }
            }
        },
        "collector.Stage": {
            "type": "object",
            "properties": {
                "col": {
                    "type": "object",
                    "$ref": "#/definitions/collector.Collector"
                },
                "lock": {
                    "type": "string"
                },
                "max": {
                    "type": "integer"
                },
                "messages": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/collector.StageMessage"
                    }
                },
                "name": {
                    "type": "string"
                },
                "percentage": {
                    "type": "number"
                },
                "progress": {
                    "type": "object",
                    "$ref": "#/definitions/collector.Progress"
                },
                "total": {
                    "type": "integer"
                }
            }
        },
        "collector.StageMessage": {
            "type": "object",
            "properties": {
                "index": {
                    "type": "integer"
                },
                "level": {
                    "type": "string"
                },
                "message": {
                    "type": "string"
                },
                "timestamp": {
                    "type": "string"
                }
            }
        },
        "controller.BulkDeleteReq": {
            "type": "object",
            "properties": {
                "names": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "controller.BulkSyncReq": {
            "type": "object",
            "properties": {
                "defs": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "controller.History": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "job": {
                    "type": "object",
                    "$ref": "#/definitions/job.FileSyncJob"
                },
                "job_id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "output": {
                    "type": "object",
                    "$ref": "#/definitions/collector.Collector"
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "controller.JobStatus": {
            "type": "object",
            "properties": {
                "definition_name": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "controller.WelcomeMessage": {
            "type": "object",
            "properties": {
                "version": {
                    "type": "string"
                }
            }
        },
        "job.FileSyncJob": {
            "type": "object",
            "properties": {
                "actionFactory": {
                    "type": "object",
                    "$ref": "#/definitions/action.Factory"
                },
                "cc": {
                    "type": "string"
                },
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "payload": {
                    "type": "object",
                    "$ref": "#/definitions/meta.FileSyncGroup"
                },
                "syncSetting": {
                    "type": "object",
                    "$ref": "#/definitions/meta.GlobalFileSyncSetting"
                }
            }
        },
        "meta.File": {
            "type": "object",
            "properties": {
                "after": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/meta.SyncAction"
                    }
                },
                "before": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/meta.SyncAction"
                    }
                },
                "dest": {
                    "type": "string"
                },
                "error": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/meta.SyncAction"
                    }
                },
                "ignores": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "skip_when_error": {
                    "type": "boolean"
                },
                "src": {
                    "type": "string"
                }
            }
        },
        "meta.FileSyncGroup": {
            "type": "object",
            "properties": {
                "after": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/meta.SyncAction"
                    }
                },
                "before": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/meta.SyncAction"
                    }
                },
                "error": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/meta.SyncAction"
                    }
                },
                "files": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/meta.File"
                    }
                },
                "from": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "rules": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/meta.Rule"
                    }
                },
                "token": {
                    "type": "string"
                }
            }
        },
        "meta.GlobalFileSyncSetting": {
            "type": "object",
            "properties": {
                "after": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/meta.SyncAction"
                    }
                },
                "before": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/meta.SyncAction"
                    }
                },
                "errors": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/meta.SyncAction"
                    }
                },
                "from": {
                    "type": "string"
                },
                "token": {
                    "type": "string"
                }
            }
        },
        "meta.Rule": {
            "type": "object",
            "properties": {
                "action": {
                    "type": "string"
                },
                "command": {
                    "type": "string"
                },
                "match": {
                    "type": "string"
                },
                "replace": {
                    "type": "string"
                },
                "src": {
                    "type": "string"
                }
            }
        },
        "meta.SyncAction": {
            "type": "object",
            "properties": {
                "action": {
                    "type": "string"
                },
                "body": {
                    "description": "--- dingding ---",
                    "type": "string"
                },
                "command": {
                    "description": "--- command ---",
                    "type": "string"
                },
                "parse_template": {
                    "type": "boolean"
                },
                "token": {
                    "type": "string"
                },
                "when": {
                    "type": "string"
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	Host:        "localhost:8819",
	BasePath:    "/api",
	Schemes:     []string{},
	Title:       "Sync API",
	Description: "文件同步服务",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
