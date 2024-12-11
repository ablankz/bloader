package cmdreqbatch

type CmdType int

const (
	_ CmdType = iota
	UpdateUserPreference
	CreateTeam
	AddUsersTeam
	CreateOrganization
	AddUsersOrganization
	CreateFileObject
	CreateTask
	UpdateStatusTask
)

func NewCmdTypeFromString(s string) CmdType {
	switch s {
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
