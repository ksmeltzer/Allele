export interface ConfigField {
  key: string;
  title?: string;
  type: string; // 'string' | 'number' | 'boolean' | 'secret' | 'enum'
  description: string;
  explanation?: string;
  defaultValue?: string;
  required: boolean;
  options?: string[];
  multiSelect?: boolean;
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
