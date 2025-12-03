import {
  DataSourceApi,
  DataSourceInstanceSettings,
  DataQueryRequest,
  DataQueryResponse,
  DataSourcePluginMeta,
} from '@grafana/data';
import { getBackendSrv, getTemplateSrv } from '@grafana/runtime';
import { Observable, from } from 'rxjs';
import { map } from 'rxjs/operators';
import { GrafanaConnectQuery, GrafanaConnectDataSourceOptions } from './types';

export class DataSource extends DataSourceApi<GrafanaConnectQuery, GrafanaConnectDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<GrafanaConnectDataSourceOptions>) {
    super(instanceSettings);
  }

  query(request: DataQueryRequest<GrafanaConnectQuery>): Observable<DataQueryResponse> {
    const { targets } = request;
    const templateSrv = getTemplateSrv();

    // Process each query
    const queries = targets
      .filter((target) => {
        // Filter out empty queries
        if (target.queryType === 'prometheus' && !target.promQL) {
          return false;
        }
        if (target.queryType === 'loki' && !target.logQL) {
          return false;
        }
        if (target.queryType === 'rest' && !target.restEndpoint) {
          return false;
        }
        return true;
      })
      .map((target) => {
        // Interpolate template variables
        const query: GrafanaConnectQuery = {
          ...target,
          promQL: target.promQL ? templateSrv.replace(target.promQL, request.scopedVars) : undefined,
          logQL: target.logQL ? templateSrv.replace(target.logQL, request.scopedVars) : undefined,
          restEndpoint: target.restEndpoint ? templateSrv.replace(target.restEndpoint, request.scopedVars) : undefined,
          restBody: target.restBody ? templateSrv.replace(target.restBody, request.scopedVars) : undefined,
        };
        return query;
      });

    if (queries.length === 0) {
      return from(Promise.resolve({ data: [] }));
    }

    // Execute queries via backend
    return from(
      getBackendSrv()
        .datasourceRequest({
          url: '/api/ds/query',
          method: 'POST',
          data: {
            queries: queries.map((q) => ({
              ...q,
              datasource: this.getRef(),
              datasourceId: this.id,
            })),
            from: request.range.from.valueOf().toString(),
            to: request.range.to.valueOf().toString(),
          },
        })
        .then((response: any) => {
          return {
            data: response.data.results ? Object.values(response.data.results).flatMap((r: any) => r.frames || []) : [],
          };
        })
        .catch((error: any) => {
          console.error('Query error:', error);
          return {
            data: [],
            error: {
              message: error.data?.message || error.message || 'Unknown error',
            },
          };
        })
    ).pipe(
      map((response) => ({
        ...response,
        data: response.data || [],
      }))
    );
  }

  async testDatasource() {
    const options = (this as any).instanceSettings.jsonData;
    
    // Check if at least one URL is configured
    if (!options.prometheusUrl && !options.lokiUrl && !options.restUrl) {
      return {
        status: 'error',
        message: 'Please configure at least one data source URL (Prometheus, Loki, or REST API).',
      };
    }

    // Try to test connectivity
    try {
      const response = await getBackendSrv().datasourceRequest({
        url: '/api/datasources/proxy/' + this.id + '/health',
        method: 'GET',
      });

      return {
        status: 'success',
        message: 'Data source is working correctly.',
      };
    } catch (error: any) {
      return {
        status: 'error',
        message: error.data?.message || error.message || 'Failed to connect to data source.',
      };
    }
  }

  getRef() {
    return {
      uid: (this as any).instanceSettings.uid,
      type: (this as any).instanceSettings.type,
    };
  }
}

