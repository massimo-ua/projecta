export class InvitationsRepository {
  #request;

  constructor(request) {
    this.#request = request;
  }

  async create(email) {
    return this.#request.post('/invitations', { email });
  }

  async list() {
    const response = await this.#request.get('/invitations');
    return response?.invitations || [];
  }

  async delete(id) {
    return this.#request.delete(`/invitations/${id}`);
  }

  async validate(code) {
    if (!code || !code.trim()) {
      throw new Error('Invitation code is required');
    }
    return this.#request.get(`/invitations/code/${encodeURIComponent(code.trim())}`, { auth: false });
  }
}

export default InvitationsRepository;
