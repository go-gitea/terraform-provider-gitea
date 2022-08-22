resource "gitea_user" "test" {
  username             = "test"
  login_name           = "test"
  password             = "Geheim1!"
  email                = "test@user.dev"
  must_change_password = false
}


resource "gitea_public_key" "test_user_key" {
  title     = "test"
  key       = file("${path.module}/id_ed25519.pub")
  username  = gitea_user.test.username
}
