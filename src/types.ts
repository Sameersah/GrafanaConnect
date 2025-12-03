import { DataQuery, DataSourceJsonData } from '@grafana/data';

export enum QueryType {
  Prometheus = 'prometheus',
  Loki = 'loki',
  REST = 'rest',
}

export interface GrafanaConnectQuery extends DataQuery {
  queryType: QueryType;
  
  // Prometheus fields
  promQL?: string;
  
  // Loki fields
  logQL?: string;
  
  // REST API fields
  restEndpoint?: string;
  restMethod?: string;
  restHeaders?: Record<string, string>;
  restBody?: string;
}

export interface GrafanaConnectDataSourceOptions extends DataSourceJsonData {
  prometheusUrl?: string;
  lokiUrl?: string;
  restUrl?: string;
  apiKey?: string;
  basicAuthUser?: string;
  bearerToken?: string;
  restHeaders?: Record<string, string>;
}

export interface GrafanaConnectSecureJsonData {
  apiKey?: string;
  basicAuthPass?: string;
  bearerToken?: string;
}

