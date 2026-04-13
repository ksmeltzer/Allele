export interface ConfigField {
  key: string;
  type: string;
  description: string;
  required: boolean;
  value?: string;
}

export interface Dependency {
  name: string;
  type: string;
  version: string;
  url: string;
}

export interface Manifest {
  name: string;
  version: string;
  description: string;
  author: string;
  dependencies: Dependency[];
  config: ConfigField[];
}
