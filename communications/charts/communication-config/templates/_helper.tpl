{{ define "root-redirect.anyshare" -}}
set $https_scheme "https";
location = / {
    if ($http_user_agent ~* (mobile|nokia|iphone|ipad|android|samsung|htc|blackberry)) {
            rewrite ^ $https_scheme://$http_host/anyshare/m/ redirect;
    }
    rewrite ^ $https_scheme://$http_host/anyshare/ redirect;
}
{{ end }}

{{ define "root-redirect.studio" -}}
set $https_scheme "https";
location = / {
    rewrite ^ $https_scheme://$http_host/studio/ redirect;
}
{{ end }}

{{ define "root-redirect.deploy" -}}
set $https_scheme "https";
location = / {
    rewrite ^ $https_scheme://$http_host/deploy/ redirect;
}
{{ end }}

{{ define "root-redirect.console" -}}
set $https_scheme "https";
location = / {
    rewrite ^ $https_scheme://$http_host/console/ redirect;
}
{{ end }}

{{ define "root-redirect.anyfabric" -}}
set $https_scheme "https";
location = / {
    rewrite ^ $https_scheme://$http_host/anyfabric/ redirect;
}
{{ end }}

