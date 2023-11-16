resource "gitea_org" "example_org" {
  name = "m_example_org"
}

resource "gitea_user" "example_users" {
  count      = 5
  username   = "m_example_user_${count.index}"
  login_name = "m_example_user_${count.index}"
  password   = "Geheim1!"
  email      = "m_example_user_${count.index}@user.dev"
}

resource "gitea_team" "example_team" {
  name         = "m_example_team"
  organisation = gitea_org.example_org.name
  description  = "An example of team membership"
  permission   = "read"
}

resource "gitea_team_members" "example_members" {
  team_id = gitea_team.example_team.id
  members = [for user in gitea_user.example_users : user.username]
}
