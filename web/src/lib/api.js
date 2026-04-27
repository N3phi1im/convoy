const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

class ApiClient {
  constructor() {
    this.baseURL = API_BASE_URL;
    this.token = localStorage.getItem('token');
  }

  setToken(token) {
    this.token = token;
    if (token) {
      localStorage.setItem('token', token);
    } else {
      localStorage.removeItem('token');
    }
  }

  getToken() {
    // Always read from localStorage to get the latest token
    return localStorage.getItem('token');
  }

  async request(endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;
    const headers = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    const token = this.getToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const config = {
      ...options,
      headers,
    };

    try {
      const response = await fetch(url, config);
      const data = await response.json();

      if (!response.ok) {
        const errorMessage = data.error?.message || data.error || data.message || 'Request failed';
        throw new Error(errorMessage);
      }

      return data.data;
    } catch (error) {
      console.error('API Error:', error);
      throw error;
    }
  }

  // Auth endpoints
  async register(email, password, displayName) {
    const data = await this.request('/auth/register', {
      method: 'POST',
      body: JSON.stringify({
        email,
        password,
        display_name: displayName,
      }),
    });
    this.setToken(data.token);
    return data;
  }

  async login(email, password) {
    const data = await this.request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({
        email,
        password,
      }),
    });
    this.setToken(data.token);
    return data;
  }

  logout() {
    this.setToken(null);
  }

  async getCurrentUser() {
    return this.request('/users/me');
  }

  // Route endpoints
  async getRoutes() {
    return this.request('/routes');
  }

  async getRoute(id) {
    return this.request(`/routes/${id}`);
  }

  async createRoute(routeData) {
    return this.request('/routes', {
      method: 'POST',
      body: JSON.stringify(routeData),
    });
  }

  async updateRoute(id, routeData) {
    return this.request(`/routes/${id}`, {
      method: 'PUT',
      body: JSON.stringify(routeData),
    });
  }

  async deleteRoute(id) {
    return this.request(`/routes/${id}`, {
      method: 'DELETE',
    });
  }

  async joinRoute(id) {
    return this.request(`/routes/${id}/join`, {
      method: 'POST',
    });
  }

  async leaveRoute(id) {
    return this.request(`/routes/${id}/leave`, {
      method: 'POST',
    });
  }

  async getRouteParticipants(id) {
    return this.request(`/routes/${id}/participants`);
  }

  async removeParticipant(routeId, userId) {
    return this.request(`/routes/${routeId}/participants/${userId}`, {
      method: 'DELETE',
    });
  }
}

export const api = new ApiClient();
