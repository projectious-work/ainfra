resource "aws_security_group" "unsafe_fixture" {
  name = "unsafe-fixture"

  ingress {
    description = "Deliberately unsafe negative scanner fixture"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
