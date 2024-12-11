package batchexecmap

type ExecType int

const (
	_ ExecType = iota
	FindJob
	FindOrganization
	GetOrganizations
	FindTask
	GetTasks
	FindFileObject
	GetFileObjects
	FindTeam
	GetTeams
	FindUserPreference
	FindUser
	GetUsers

	CreateUserProfile
	UpdateUserPreference
	CreateTeam
	AddUsersTeam
	CreateOrganization
	AddUsersOrganization
	CreateFileObject
	CreateTask
	UpdateStatusTask
)

func NewExecTypeFromString(s string) ExecType {
	switch s {
	case "FindJob":
		return FindJob
	case "FindOrganization":
		return FindOrganization
	case "GetOrganizations":
		return GetOrganizations
	case "FindTask":
		return FindTask
	case "GetTasks":
		return GetTasks
	case "FindFileObject":
		return FindFileObject
	case "GetFileObjects":
		return GetFileObjects
	case "FindTeam":
		return FindTeam
	case "GetTeams":
		return GetTeams
	case "FindUserPreference":
		return FindUserPreference
	case "FindUser":
		return FindUser
	case "GetUsers":
		return GetUsers
	case "CreateUserProfile":
		return CreateUserProfile
	case "UpdateUserPreference":
		return UpdateUserPreference
	case "CreateTeam":
		return CreateTeam
	case "AddUsersTeam":
		return AddUsersTeam
	case "CreateOrganization":
		return CreateOrganization
	case "AddUsersOrganization":
		return AddUsersOrganization
	case "CreateFileObject":
		return CreateFileObject
	case "CreateTask":
		return CreateTask
	case "UpdateStatusTask":
		return UpdateStatusTask
	default:
		return 0
	}
}
