package router

import (
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_createPathMatcher(t *testing.T) {
	tests := []struct {
		path         string
		wantRegex    string
		wantCaptures []string
	}{
		{
			path:      `/blog/:year/:month`,
			wantRegex: `/blog/([^/]+)/([^/]+)`,
			wantCaptures: []string{
				"year",
				"month",
			},
		},
		{
			path:      `/blog/{year}/{month}`,
			wantRegex: `/blog/([^/]+)/([^/]+)`,
			wantCaptures: []string{
				"year",
				"month",
			},
		},
		{
			path:      `/say/*/to/*`,
			wantRegex: `/say/(.+)/to/(.+)`,
			wantCaptures: []string{
				wildcardKey,
				wildcardKey,
			},
		},
		{
			path:      `/download/*.*`,
			wantRegex: `/download/(.+)\.(.+)`,
			wantCaptures: []string{
				wildcardKey,
				wildcardKey,
			},
		},
		{
			path:      `/blog/{year:\d{4}}`,
			wantRegex: `/blog/(\d{4})`,
			wantCaptures: []string{
				"year",
			},
		},
		{
			path:      `/hi/{user:.*}`,
			wantRegex: `/hi/(.*)`,
			wantCaptures: []string{
				"user",
			},
		},
		{
			path:      `/blog/{year:(?:199\d|20\d{2})}/{month:(?:0?[1-9]|1[0-2])}`,
			wantRegex: `/blog/((?:199\d|20\d{2}))/((?:0?[1-9]|1[0-2]))`,
			wantCaptures: []string{
				"year",
				"month",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			gotRegex, gotCaptures := createPathMatcher(tt.path)
			if gotRegex != tt.wantRegex {
				t.Errorf("createPathMatcher() gotRegex = %v, want %v", gotRegex, tt.wantRegex)
			}
			if diff := cmp.Diff(tt.wantCaptures, gotCaptures); diff != "" {
				t.Errorf("createPathMatcher() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_regexpCapture_MatchPath(t *testing.T) {
	type fields struct {
		rg       *regexp.Regexp
		captures []string
	}
	tests := []struct {
		name    string
		fields  fields
		path    string
		want    *params
		wantErr bool
	}{
		{
			name: "/blog/([^/]+)/([^/]+)",
			fields: fields{
				rg: regexp.MustCompile(`/blog/([^/]+)/([^/]+)`),
				captures: []string{
					"year",
					"month",
				},
			},
			path: "/blog/2018/01",
			want: &params{
				wildcards: []string{},
				capture: map[string]string{
					"year":  "2018",
					"month": "01",
				},
			},
		},
		{
			name: "/say/(.+)/to/(.+)",
			fields: fields{
				rg: regexp.MustCompile(`/say/(.+)/to/(.+)`),
				captures: []string{
					wildcardKey,
					wildcardKey,
				},
			},
			path: "/say/foo/to/bar",
			want: &params{
				wildcards: []string{
					"foo",
					"bar",
				},
				capture: map[string]string{},
			},
		},
		{
			name: `/download/(.+)\.(.+)`,
			fields: fields{
				rg: regexp.MustCompile(`/download/(.+)\.(.+)`),
				captures: []string{
					wildcardKey,
					wildcardKey,
				},
			},
			path: "/download/path/to/file.xml",
			want: &params{
				wildcards: []string{
					"path/to/file",
					"xml",
				},
				capture: map[string]string{},
			},
		},
		{
			name: `/blog/(\d{4})`,
			fields: fields{
				rg: regexp.MustCompile(`/blog/(\d{4})`),
				captures: []string{
					"year",
				},
			},
			path: "/blog/2018",
			want: &params{
				wildcards: []string{},
				capture: map[string]string{
					"year": "2018",
				},
			},
		},
		{
			name: `/hi/(.*)`,
			fields: fields{
				rg: regexp.MustCompile(`/hi/(.*)`),
				captures: []string{
					"user",
				},
			},
			path: "/hi/codehex",
			want: &params{
				wildcards: []string{},
				capture: map[string]string{
					"user": "codehex",
				},
			},
		},
		{
			name: `/blog/((?:199\d|20\d{2}))/((?:0?[1-9]|1[0-2]))`,
			fields: fields{
				rg: regexp.MustCompile(`/blog/((?:199\d|20\d{2}))/((?:0?[1-9]|1[0-2]))`),
				captures: []string{
					"year",
					"month",
				},
			},
			path: "/blog/2018/01",
			want: &params{
				wildcards: []string{},
				capture: map[string]string{
					"year":  "2018",
					"month": "01",
				},
			},
		},
		{
			name: "mismatch numbers with captures",
			fields: fields{
				rg: regexp.MustCompile(`/blog/([^/]+)/([^/]+)`), // 2 group
				captures: []string{
					"year", // but 1 capture
				},
			},
			path:    "/blog/2018/01",
			wantErr: true,
		},
		{
			name: "did not match",
			fields: fields{
				rg: regexp.MustCompile(`/blog/([^/]+)/([^/]+)`), // 2 group
			},
			path:    "/foo/bar",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := regexpCapture{
				rg:       tt.fields.rg,
				captures: tt.fields.captures,
			}
			got, err := rc.MatchPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("regexpCapture.MatchPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(params{})); diff != "" {
				t.Errorf("regexpCapture.MatchPath() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
