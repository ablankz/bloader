package batchexecmap

import (
	"context"

	"github.com/LabGroupware/go-measure-tui/internal/api/request/executor"
	"github.com/LabGroupware/go-measure-tui/internal/app"
	"github.com/LabGroupware/go-measure-tui/internal/auth"
	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/cmdreqbatch"
	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/execbatch"
	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/queryreqbatch"
)

type ExecutorFactory interface {
	Factory(
		ctx context.Context,
		ctr *app.Container,
		id int,
		request *execbatch.ValidatedExecRequest,
		termChan chan<- execbatch.TerminateType,
		authToken *auth.AuthToken,
		apiEndpoint string,
		consumer execbatch.ResponseDataConsumer,
	) (executor.RequestExecutor, func(), error)
}

var TypeFactoryMap = map[ExecType]ExecutorFactory{
	FindJob:            queryreqbatch.FindJobFactory{},
	FindTask:           queryreqbatch.FindTaskFactory{},
	FindTeam:           queryreqbatch.FindTeamFactory{},
	FindUserPreference: queryreqbatch.FindUserPreferenceFactory{},
	FindFileObject:     queryreqbatch.FindFileObjectFactory{},
	FindOrganization:   queryreqbatch.FindOrganizationFactory{},
	FindUser:           queryreqbatch.FindUserFactory{},
	GetTasks:           queryreqbatch.GetTasksFactory{},
	GetTeams:           queryreqbatch.GetTeamsFactory{},
	GetFileObjects:     queryreqbatch.GetFileObjectsFactory{},
	GetOrganizations:   queryreqbatch.GetOrganizationsFactory{},
	GetUsers:           queryreqbatch.GetUsersFactory{},

	CreateUserProfile:    cmdreqbatch.CreateUserProfileFactory{},
	UpdateUserPreference: cmdreqbatch.UpdateUserPreferenceFactory{},
	CreateTeam:           cmdreqbatch.CreateTeamFactory{},
	AddUsersTeam:         cmdreqbatch.AddUsersTeamFactory{},
	CreateOrganization:   cmdreqbatch.CreateOrganizationFactory{},
	AddUsersOrganization: cmdreqbatch.AddUsersOrganizationFactory{},
	CreateTask:           cmdreqbatch.CreateTaskFactory{},
	UpdateStatusTask:     cmdreqbatch.UpdateStatusTaskFactory{},
	CreateFileObject:     cmdreqbatch.CreateFileObjectFactory{},
}
