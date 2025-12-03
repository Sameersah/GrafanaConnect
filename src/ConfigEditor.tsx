import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { GrafanaConnectDataSourceOptions, GrafanaConnectSecureJsonData } from './types';

const { FormField, SecretFormField } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<GrafanaConnectDataSourceOptions, GrafanaConnectSecureJsonData> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onPrometheusUrlChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      prometheusUrl: (event.target as HTMLInputElement).value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onLokiUrlChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      lokiUrl: (event.target as HTMLInputElement).value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onRESTUrlChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      restUrl: (event.target as HTMLInputElement).value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onAPIKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      apiKey: (event.target as HTMLInputElement).value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onAPIKeySecretChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...options.secureJsonData,
        apiKey: (event.target as HTMLInputElement).value,
      },
    });
  };

  onAPIKeySecretReset = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        apiKey: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        apiKey: '',
      },
    });
  };

  onBasicAuthUserChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      basicAuthUser: (event.target as HTMLInputElement).value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onBasicAuthPassChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...options.secureJsonData,
        basicAuthPass: (event.target as HTMLInputElement).value,
      },
    });
  };

  onBasicAuthPassReset = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        basicAuthPass: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        basicAuthPass: '',
      },
    });
  };

  onBearerTokenChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      bearerToken: (event.target as HTMLInputElement).value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onBearerTokenSecretChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...options.secureJsonData,
        bearerToken: (event.target as HTMLInputElement).value,
      },
    });
  };

  onBearerTokenSecretReset = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        bearerToken: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        bearerToken: '',
      },
    });
  };

  render() {
    const { options } = this.props;
    const { jsonData, secureJsonData, secureJsonFields } = options;

    return (
      <div className="gf-form-group">
        <div className="gf-form">
          <h3>Data Source URLs</h3>
        </div>

        <div className="gf-form">
          <FormField
            label="Prometheus URL"
            labelWidth={10}
            inputWidth={20}
            onChange={this.onPrometheusUrlChange}
            value={jsonData.prometheusUrl || ''}
            placeholder="http://prometheus:9090"
            tooltip="Base URL for Prometheus instance"
          />
        </div>

        <div className="gf-form">
          <FormField
            label="Loki URL"
            labelWidth={10}
            inputWidth={20}
            onChange={this.onLokiUrlChange}
            value={jsonData.lokiUrl || ''}
            placeholder="http://loki:3100"
            tooltip="Base URL for Loki instance"
          />
        </div>

        <div className="gf-form">
          <FormField
            label="REST API Base URL"
            labelWidth={10}
            inputWidth={20}
            onChange={this.onRESTUrlChange}
            value={jsonData.restUrl || ''}
            placeholder="https://api.example.com"
            tooltip="Base URL for REST API endpoints"
          />
        </div>

        <div className="gf-form">
          <h3>Authentication</h3>
        </div>

        <div className="gf-form">
          <FormField
            label="API Key (Plain)"
            labelWidth={10}
            inputWidth={20}
            onChange={this.onAPIKeyChange}
            value={jsonData.apiKey || ''}
            placeholder="Optional API key"
            tooltip="API key for authentication (stored in plain text)"
          />
        </div>

        <div className="gf-form">
          <SecretFormField
            isConfigured={secureJsonFields?.apiKey}
            value={secureJsonData?.apiKey || ''}
            label="API Key (Secure)"
            labelWidth={10}
            inputWidth={20}
            onReset={this.onAPIKeySecretReset}
            onChange={this.onAPIKeySecretChange}
            placeholder="Optional secure API key"
            tooltip="API key stored securely"
          />
        </div>

        <div className="gf-form">
          <FormField
            label="Basic Auth Username"
            labelWidth={10}
            inputWidth={20}
            onChange={this.onBasicAuthUserChange}
            value={jsonData.basicAuthUser || ''}
            placeholder="Optional username"
            tooltip="Username for basic authentication"
          />
        </div>

        <div className="gf-form">
          <SecretFormField
            isConfigured={secureJsonFields?.basicAuthPass}
            value={secureJsonData?.basicAuthPass || ''}
            label="Basic Auth Password"
            labelWidth={10}
            inputWidth={20}
            onReset={this.onBasicAuthPassReset}
            onChange={this.onBasicAuthPassChange}
            placeholder="Optional password"
            tooltip="Password for basic authentication (stored securely)"
          />
        </div>

        <div className="gf-form">
          <FormField
            label="Bearer Token (Plain)"
            labelWidth={10}
            inputWidth={20}
            onChange={this.onBearerTokenChange}
            value={jsonData.bearerToken || ''}
            placeholder="Optional bearer token"
            tooltip="Bearer token for authentication (stored in plain text)"
          />
        </div>

        <div className="gf-form">
          <SecretFormField
            isConfigured={secureJsonFields?.bearerToken}
            value={secureJsonData?.bearerToken || ''}
            label="Bearer Token (Secure)"
            labelWidth={10}
            inputWidth={20}
            onReset={this.onBearerTokenSecretReset}
            onChange={this.onBearerTokenSecretChange}
            placeholder="Optional secure bearer token"
            tooltip="Bearer token stored securely"
          />
        </div>
      </div>
    );
  }
}

