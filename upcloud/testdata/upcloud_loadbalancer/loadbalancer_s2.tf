variable "basename" {
  default = "tf-acc-test-lb-"
  type    = string
}

variable "zone" {
  type    = string
  default = "fi-hel2"
}

resource "upcloud_network" "lb_network" {
  name = "${var.basename}net"
  zone = var.zone
  ip_network {
    address = "10.0.7.0/24"
    dhcp    = true
    family  = "IPv4"
  }
}

resource "upcloud_floating_ip_address" "ip" {
  count = 3

  access         = "public"
  family         = "IPv4"
  release_policy = "keep"
  zone           = var.zone
}

resource "upcloud_loadbalancer" "lb" {
  configured_status = "started"
  name              = "${var.basename}lb"
  plan              = "production-small"
  zone              = var.zone
  maintenance_dow   = "monday"
  maintenance_time  = "00:01:01Z"

  // Attach 1 new floating IP address
  ip_addresses = [
    for ip in slice(upcloud_floating_ip_address.ip, 0, 2) :
    {
      address      = ip.ip_address
      network_name = "public"
    }
  ]

  networks {
    type   = "public"
    family = "IPv4"
    name   = "public"
  }

  networks {
    type    = "private"
    family  = "IPv4"
    name    = "private"
    network = resource.upcloud_network.lb_network.id
  }

  labels = {
    key       = "value"
    test-step = "2"
  }
}

resource "upcloud_loadbalancer_frontend" "lb_fe_1" {
  loadbalancer = resource.upcloud_loadbalancer.lb.id
  name         = "lb-fe-1-test"
  mode         = "http"
  port         = 8080
  # change(indirect): backend lb_be_1 name
  default_backend_name = resource.upcloud_loadbalancer_backend.lb_be_1.name
  properties {
    timeout_client         = 20
    inbound_proxy_protocol = true
  }

  networks {
    name = "public"
  }
}

resource "upcloud_loadbalancer_frontend" "lb_fe_2" {
  loadbalancer         = resource.upcloud_loadbalancer.lb.id
  name                 = "lb-fe-2-test"
  mode                 = "http"
  port                 = 9090
  default_backend_name = resource.upcloud_loadbalancer_backend.lb_be_2.name

  networks {
    name = "public"
  }

  properties {}
}

resource "upcloud_loadbalancer_resolver" "lb_dns_1" {
  loadbalancer = resource.upcloud_loadbalancer.lb.id
  # change: name lb-resolver-1-test to lb-resolver-1-test-1
  name          = "lb-resolver-1-test-1"
  cache_invalid = 10
  cache_valid   = 100
  retries       = 5
  timeout       = 10
  timeout_retry = 10
  nameservers   = ["94.237.127.9:53", "94.237.40.9:53"]
}

resource "upcloud_loadbalancer_backend" "lb_be_1" {
  loadbalancer = resource.upcloud_loadbalancer.lb.id
  # change(indirect): resolver name changed
  resolver_name = resource.upcloud_loadbalancer_resolver.lb_dns_1.name
  # change: name lb-be-1-test to lb-be-1-test-1
  name = "lb-be-1-test-1"
  properties {
    timeout_server          = 20
    timeout_tunnel          = 4000
    health_check_type       = "http"
    outbound_proxy_protocol = "v1"
    sticky_session_cookie_name = "Sticky-Session"
  }
}

resource "upcloud_loadbalancer_static_backend_member" "lb_be_1_sm_1" {
  backend      = resource.upcloud_loadbalancer_backend.lb_be_1.id
  name         = "lb-be-1-sm-1-test"
  ip           = "10.0.0.10"
  port         = 8000
  weight       = 100
  max_sessions = 1000
  enabled      = true
}

resource "upcloud_loadbalancer_dynamic_backend_member" "lb_be_1_dm_1" {
  backend      = resource.upcloud_loadbalancer_backend.lb_be_1.id
  name         = "lb-be-1-dm-1-test"
  weight       = 10
  max_sessions = 10
  enabled      = false
}

resource "upcloud_loadbalancer_backend" "lb_be_2" {
  loadbalancer = resource.upcloud_loadbalancer.lb.id
  name         = "lb-be-2-test"

  properties {}
}

resource "upcloud_loadbalancer_static_backend_member" "lb_be_2_sm_1" {
  backend = resource.upcloud_loadbalancer_backend.lb_be_2.id
  name    = "lb-be-2-sm-1-test"
  ip      = "10.0.0.10"
  port    = 8000
  # set weight to zero
  weight       = 0
  max_sessions = 0
  enabled      = true
}

resource "upcloud_loadbalancer_frontend_rule" "lb_fe_1_r1" {
  frontend = resource.upcloud_loadbalancer_frontend.lb_fe_1.id
  name     = "lb-fe-1-r1-test"
  priority = 10

  # change: truncate matchers
  matchers {
    src_port {
      method = "equal"
      value  = 80
    }
  }
  # change: truncate actions
  actions {
    use_backend {
      backend_name = resource.upcloud_loadbalancer_backend.lb_be_1.name
    }
  }
}

resource "upcloud_loadbalancer_dynamic_certificate_bundle" "lb_cb_d1" {
  name = "${var.basename}dynamic-cert"
  hostnames = [
    "example.com",
  ]
  key_type = "rsa"
}

resource "upcloud_loadbalancer_manual_certificate_bundle" "lb_cb_m1" {
  name        = "${var.basename}manual-cert"
  certificate = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURRVENDQWltZ0F3SUJBZ0lCQVRBTkJna3Foa2lHOXcwQkFRc0ZBREJDTVFzd0NRWURWUVFHRXdKWVdERVYKTUJNR0ExVUVCd3dNUkdWbVlYVnNkQ0JEYVhSNU1Sd3dHZ1lEVlFRS0RCTkVaV1poZFd4MElFTnZiWEJoYm5rZwpUSFJrTUI0WERUSTFNVEl4TVRFeE1ETXpNVm9YRFRJMk1USXhNVEV4TURNek1Wb3dRakVMTUFrR0ExVUVCaE1DCldGZ3hGVEFUQmdOVkJBY01ERVJsWm1GMWJIUWdRMmwwZVRFY01Cb0dBMVVFQ2d3VFJHVm1ZWFZzZENCRGIyMXcKWVc1NUlFeDBaRENDQVNJd0RRWUpLb1pJaHZjTkFRRUJCUUFEZ2dFUEFEQ0NBUW9DZ2dFQkFNMmI4QjZRRUliUQp0eER4aktkdkY5VWhvbSt2am1tK29nUDdxRUIycHE2N3FWQUhzcWtSSm9BUmF3VHhub1g4SGsxY1YxRW1TQmVwCmhkMnA2SkxVYmhPcS9aaU45NVN4QXpwUWxtMk1GUzE0Q3Q2NlFIMisySVI1allSR2lRQTBobjN6N3ZyVWF3NGEKZk9BTlJueWlCcUJpcHZFOUc2U1RHTTlVeThub292Qmg4N29sSGNUVE00RnRuYUNWMnA0UlIweFpwa0VzS2UrcAoyMGFNdDJMYWJrTkRFcC9BNmQrUzljc0txNnpGbkxBdHU0c1FLWU5lcXlmbDcxZXowQ0MxV3I4ZExwWTZLb20yCjNWMmtUdlJuSjZibk9YK1JsendDbzJyT3pFclBkZTM4Q2F2aVZYcllLUHhuMDJtakoxMW0zNGFWVmFiWXJtVG8KaGxDMW9rcmwzQmtDQXdFQUFhTkNNRUF3SFFZRFZSME9CQllFRklpcWZTQU1VZ2M0dUttZ0NMQ09nWVlqNmowQgpNQjhHQTFVZEl3UVlNQmFBRlBaVEQ3cExpcFJucUJ6WWEvZzA4OVlNelBwa01BMEdDU3FHU0liM0RRRUJDd1VBCkE0SUJBUUJEKzE4NTU4SWttOEZRUy91V25icUFrU0Q0Tm1tM1R5bFdNa3ZiSGRrejZzZ0FTOEZjMUI5b2ZORm4KR09WRDdXNkE3Vkx1L3VZTFE3b3k3NXBub1VPdVNQV1hWeEtHM0k5TUFpNlA4V2FWZmNUS05HSkFSR25CY1cvdQp4S29HSVY5dC9WRzNHWUVuQzFJN1FSR2dsQy9SeGFEWVE5Tnc1bkM2OEMrZ1JtalVhdEpPdDNLVmR3U3luUnAxCno4b0tEaVBFbjFHT0pYRm54eWN6Y0lydE5FdFd3ZGZEN24yTS90OEVHaUQ4eGJJdnVKVXVmdGRucFJhN3ZJNTcKYVJvUHRmenh6bUg4YTFQTWxXRnRldEgvbFFMLzdHNWdiZU4xd2VRczJjZzJ1bHROTmltajVmNmNHNUhYODM2RAo0QWdJNWFjMDBlalpVYWpycFRDVVRnRThENXpRCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
  private_key = "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2d0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktrd2dnU2xBZ0VBQW9JQkFRRE5tL0Fla0JDRzBMY1EKOFl5bmJ4ZlZJYUp2cjQ1cHZxSUQrNmhBZHFhdXU2bFFCN0twRVNhQUVXc0U4WjZGL0I1TlhGZFJKa2dYcVlYZApxZWlTMUc0VHF2MllqZmVVc1FNNlVKWnRqQlV0ZUFyZXVrQjl2dGlFZVkyRVJva0FOSVo5OCs3NjFHc09HbnpnCkRVWjhvZ2FnWXFieFBSdWtreGpQVk12SjZLTHdZZk82SlIzRTB6T0JiWjJnbGRxZUVVZE1XYVpCTENudnFkdEcKakxkaTJtNURReEtmd09uZmt2WExDcXVzeFp5d0xidUxFQ21EWHFzbjVlOVhzOUFndFZxL0hTNldPaXFKdHQxZApwRTcwWnllbTV6bC9rWmM4QXFOcXpzeEt6M1h0L0FtcjRsVjYyQ2o4WjlOcG95ZGRadCtHbFZXbTJLNWs2SVpRCnRhSks1ZHdaQWdNQkFBRUNnZ0VBUEZWWTVhOENtbnplYTBObU1hK2d2N0xwOW5uK2dUc21VYUxrSVY1djFQQk8KWTZTT29admR2MURkSllzOUtEWHVNbWM1WEIrdW9mcmx4RURhZFZPT3BZalVkNUtaSnZHMmI4TThFUk05RjZXVgpFdngyZGkrdFcxcEwwNWZiRmN0VDk5dS9zYXpwYVM4T203UnBqYU1CN01obUVuNExBWVVFajdwalBuRmNkc3JRClE0cE9NenhRTk5zM1VtRThibXZXV0ZkeGF1Yi9yVkV1VU9zMWNiNmU0L2hOTkJHNjBTOGpFRVpuQU1QSm9BK3IKbGRmd2lGYnZuMXhLcjZaTWZmU0IvOXU0bGUrSlRSS2ZzV2U0bzZBS2R2Z3ZkY3BYcmRtMlY1UTNYZ1BqYlNWWApPTExUTXdhUjZaV3QrQnVQNExpOEFoL3ozR0tMbnlzck5pRFpSWFVkelFLQmdRRDdaM21PdFF3d3NHMjNmTDFKClhtaHd5azBKcWJwMkhWNmU0UEdWK214VkoyL0l1RkhKeDRWbGdDbkxLeU56bWxyaGNWcks1MkozYlFyaW9CRHoKcVM2Vmdpd3loeUJxUzI0RzJvN2JzdFJIbjNhZklIQ3VsZituRHM3TklBTVpRSWVaVjFINjNjcE1pZWVhNzYydwpuWVc2enRDd3BSWm9pc3F2RDRTWCtycXR6d0tCZ1FEUlhpYVR0cXlsemVmWGxRNUZGVGZCeDNKYmR5aExFUDZhCncyR0RFOWlFamROTXVLRnJSRjVqeklDL25GQXE3SjZia1YvN2Y3c3dBdzlxOTBmQXljZkNiS1F5TFJxV0lUMDAKM2JFbEdRUk4yZ0xlSitNVkZMSEVocEJXZHBEZGsxM2lObXpkcnhGeFhXRmlEVmNvM3huVUx0NmtqTDROaS9jdApuaS94Z2xYNWx3S0JnUUN3alJKSXJjeEp4UnpINXNublpHMWtDQzNodzFnMjZwa3dhamcrWXdjQkpoalNsTjZiCkhZc0lwT0MwMVM2b1dKWEtESmorTlZCcEhpS3UxRW9UVTVScldtYy9kTFhHOEFIc3ZqL2srY2txSTBwaXBaMTgKZmNwenYycHJremVaM0Q5ZDZIeWgrRy9CSUhlTnp4UGpIRHgxM0JlaWRjMHV6WWxaTjBTZWxtM1M4UUtCZ1FERgpwTGFRSFJOd1ZpZDFxTjFXczhmMTR6eitRVWRGVGQ2NzVKTnA5Tk1obHUwUWNQN1l6eXEzMVhiNDZ5djJ5WGFVCjd6Q0hyN1hhaGhrSTVqVFROdWlmam9XV1pHUERzODhlMStVQlcxTm4xdFY4T0hVekVsMGFZOWxmOWYrZFhCOTEKaStGTGlKZlR4ODVGak1ocDZlcHRGbTNSTXBlN0hCVVQrRS9VRWpEdE13S0JnUUNPNEppV0NScVBweXhTTkFFegp2NHhUUmsyRHJLOWdKTVFqU1dsTWlQUldkalpvTnkzMTdsS2c2V3JBTGRiVU9DOUEzUlg4NnpDU00zT0tuTW91Cmt2TlJMTnBuTEdWbkZMN0U4MWdkc1pYazB5Tzg2bTg1UW9RcFdPZkxrMnBsdUg4ZU02VDNqUnBVcHFWTlV6TXYKZlRoYVJiM0VqTmJVRUx2S3R5V1NnZG9rd1E9PQotLS0tLUVORCBQUklWQVRFIEtFWS0tLS0tCg=="
}

resource "upcloud_loadbalancer_frontend_tls_config" "lb_fe_1_tls1" {
  frontend           = resource.upcloud_loadbalancer_frontend.lb_fe_1.id
  name               = "lb-fe-1-tls1-test"
  certificate_bundle = resource.upcloud_loadbalancer_manual_certificate_bundle.lb_cb_m1.id
}

resource "upcloud_loadbalancer_backend_tls_config" "lb_be_1_tls1" {
  backend            = resource.upcloud_loadbalancer_backend.lb_be_1.id
  name               = "lb-be-1-tls1-test"
  certificate_bundle = resource.upcloud_loadbalancer_manual_certificate_bundle.lb_cb_m1.id
}
