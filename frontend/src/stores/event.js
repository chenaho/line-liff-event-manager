import { defineStore } from 'pinia'
import axios from 'axios'

export const useEventStore = defineStore('event', {
    state: () => ({
        events: [],
        currentEvent: null,
        loading: false,
        error: null
    }),
    actions: {
        async fetchEvents() {
            this.loading = true
            try {
                const response = await axios.get('/api/events')
                // Ensure events is always an array
                this.events = response.data || []
            } catch (err) {
                this.error = 'Fetch Events Failed: ' + err.message
                console.error(err)
                // Keep events as empty array on error
                this.events = []
            } finally {
                this.loading = false
            }
        },
        async createEvent(eventData) {
            this.loading = true
            try {
                const response = await axios.post('/api/events', eventData)
                // Ensure events is an array before using unshift
                if (!this.events) {
                    this.events = []
                }
                this.events.unshift(response.data)
                return response.data
            } catch (err) {
                this.error = 'Create Event Failed: ' + err.message
                console.error(err)
                throw err
            } finally {
                this.loading = false
            }
        },
        async updateEventStatus(eventId, isActive) {
            try {
                await axios.put(`/api/events/${eventId}/status`, { isActive })
                const event = this.events.find(e => e.eventId === eventId)
                if (event) {
                    event.isActive = isActive
                }
            } catch (err) {
                this.error = 'Update Status Failed: ' + err.message
                console.error(err)
            }
        },
        async fetchEvent(eventId) {
            this.loading = true
            try {
                const response = await axios.get(`/api/events/${eventId}`)
                this.currentEvent = response.data
                return response.data
            } catch (err) {
                this.error = 'Fetch Event Failed: ' + err.message
                console.error(err)
            } finally {
                this.loading = false
            }
        },
        async submitAction(eventId, type, payload) {
            try {
                await axios.post(`/api/events/${eventId}/action`, { type, payload })
                // Refresh status after action
                await this.fetchEventStatus(eventId)
            } catch (err) {
                this.error = 'Action Failed: ' + err.message
                console.error(err)
                throw err
            }
        },
        async fetchEventStatus(eventId) {
            try {
                const response = await axios.get(`/api/events/${eventId}/status`)
                // Merge status into currentEvent or store separately
                // For simplicity, let's return it or store in a separate state
                return response.data
            } catch (err) {
                console.error(err)
            }
        }
    }
})

