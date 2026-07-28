import React from 'react';
import PropTypes from 'prop-types';
import { ErrorPage } from './ErrorPage';

export class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    console.error('Uncaught error in ErrorBoundary:', error, errorInfo);
    const { onError } = this.props;
    if (onError) {
      onError(error, errorInfo);
    }
  }

  componentDidUpdate(prevProps) {
    const { resetKeys } = this.props;
    const { hasError } = this.state;
    if (hasError && resetKeys && prevProps.resetKeys) {
      if (resetKeys.some((key, index) => key !== prevProps.resetKeys[index])) {
        this.resetErrorBoundary();
      }
    }
  }

  resetErrorBoundary = () => {
    const { onReset } = this.props;
    if (onReset) {
      onReset();
    }
    this.setState({ hasError: false, error: null });
  };

  render() {
    const { hasError, error } = this.state;
    const { fallback, children } = this.props;

    if (hasError) {
      if (fallback) {
        if (typeof fallback === 'function') {
          return fallback({
            error,
            resetErrorBoundary: this.resetErrorBoundary,
          });
        }
        return fallback;
      }
      return (
        <ErrorPage
          error={error}
          resetErrorBoundary={this.resetErrorBoundary}
        />
      );
    }

    return children;
  }
}

ErrorBoundary.propTypes = {
  children: PropTypes.node.isRequired,
  fallback: PropTypes.oneOfType([PropTypes.node, PropTypes.func]),
  onError: PropTypes.func,
  onReset: PropTypes.func,
  resetKeys: PropTypes.arrayOf(PropTypes.any),
};

ErrorBoundary.defaultProps = {
  fallback: null,
  onError: null,
  onReset: null,
  resetKeys: null,
};

export default ErrorBoundary;
