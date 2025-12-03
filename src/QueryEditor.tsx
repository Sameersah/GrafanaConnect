import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { GrafanaConnectQuery, QueryType } from './types';

const { FormField, Select } = LegacyForms;

interface Props extends QueryEditorProps<any, GrafanaConnectQuery> {}

interface State {}

const queryTypeOptions = [
  { value: QueryType.Prometheus, label: 'Prometheus' },
  { value: QueryType.Loki, label: 'Loki' },
  { value: QueryType.REST, label: 'REST API' },
];

const httpMethodOptions = [
  { value: 'GET', label: 'GET' },
  { value: 'POST', label: 'POST' },
  { value: 'PUT', label: 'PUT' },
  { value: 'PATCH', label: 'PATCH' },
  { value: 'DELETE', label: 'DELETE' },
];

export class QueryEditor extends PureComponent<Props, State> {
  onQueryTypeChange = (option: any) => {
    const { onChange, query } = this.props;
    onChange({
      ...query,
      queryType: option.value,
    });
  };

  onPromQLChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({
      ...query,
      promQL: event.target.value,
    });
  };

  onLogQLChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({
      ...query,
      logQL: event.target.value,
    });
  };

  onRESTEndpointChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({
      ...query,
      restEndpoint: event.target.value,
    });
  };

  onRESTMethodChange = (option: any) => {
    const { onChange, query } = this.props;
    onChange({
      ...query,
      restMethod: option.value,
    });
  };

  onRESTBodyChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    const { onChange, query } = this.props;
    onChange({
      ...query,
      restBody: event.target.value,
    });
  };

  renderPrometheusEditor() {
    const { query } = this.props;
    return (
      <div className="gf-form">
        <FormField
          label="PromQL Query"
          labelWidth={10}
          inputWidth={20}
          onChange={this.onPromQLChange}
          value={query.promQL || ''}
          placeholder="up{job=\"prometheus\"}"
          tooltip="Enter a PromQL query expression"
        />
      </div>
    );
  }

  renderLokiEditor() {
    const { query } = this.props;
    return (
      <div className="gf-form">
        <FormField
          label="LogQL Query"
          labelWidth={10}
          inputWidth={20}
          onChange={this.onLogQLChange}
          value={query.logQL || ''}
          placeholder="{job=\"varlogs\"}"
          tooltip="Enter a LogQL query expression"
        />
      </div>
    );
  }

  renderRESTEditor() {
    const { query } = this.props;
    return (
      <>
        <div className="gf-form">
          <FormField
            label="Endpoint"
            labelWidth={10}
            inputWidth={20}
            onChange={this.onRESTEndpointChange}
            value={query.restEndpoint || ''}
            placeholder="/api/v1/metrics"
            tooltip="REST API endpoint path (relative to base URL)"
          />
        </div>
        <div className="gf-form">
          <Select
            label="HTTP Method"
            labelWidth={10}
            width={20}
            options={httpMethodOptions}
            value={httpMethodOptions.find((o) => o.value === (query.restMethod || 'GET'))}
            onChange={this.onRESTMethodChange}
            tooltip="HTTP method for the request"
          />
        </div>
        {(query.restMethod === 'POST' || query.restMethod === 'PUT' || query.restMethod === 'PATCH') && (
          <div className="gf-form">
            <label className="gf-form-label width-10">Request Body</label>
            <textarea
              className="gf-form-input width-20"
              rows={5}
              onChange={this.onRESTBodyChange}
              value={query.restBody || ''}
              placeholder='{"key": "value"}'
            />
          </div>
        )}
      </>
    );
  }

  render() {
    const { query } = this.props;
    const queryType = query.queryType || QueryType.Prometheus;

    return (
      <div className="gf-form-group">
        <div className="gf-form">
          <Select
            label="Query Type"
            labelWidth={10}
            width={20}
            options={queryTypeOptions}
            value={queryTypeOptions.find((o) => o.value === queryType)}
            onChange={this.onQueryTypeChange}
            tooltip="Select the type of data source to query"
          />
        </div>

        {queryType === QueryType.Prometheus && this.renderPrometheusEditor()}
        {queryType === QueryType.Loki && this.renderLokiEditor()}
        {queryType === QueryType.REST && this.renderRESTEditor()}
      </div>
    );
  }
}

