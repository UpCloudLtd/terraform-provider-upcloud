($m[0] | {full_matrix, groups}) as $cfg
| (
    def full_matrix:
      ($cfg.groups | keys) + ["unittests"] | unique | sort;

    def rule_match($path; $rule):
      if ($rule | endswith("/")) then ($path | startswith($rule)) else $path == $rule end;

    def is_full_path($path):
      ($cfg.full_matrix // []) | any(rule_match($path; .));

    def groups_hit($path):
      [ ($cfg.groups | to_entries)[]
        | select(.value | any(rule_match($path; .)))
        | .key ]
      | unique;

    def partial:
      ["unittests"]
      + (($paths | map(groups_hit(.)) | add // []) | unique | map(select(. != "unittests")) | sort);

    if ($paths | length) == 0 then full_matrix
    elif any($paths[]; is_full_path(.)) then full_matrix
    elif ($paths | all((groups_hit(.) | length) == 0)) then ["unittests"]
    else partial
    end
  )

