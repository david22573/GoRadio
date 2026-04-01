export const load = async ({ fetch, params }) => {
	try {
		const res = await fetch('/api/stations');
		const data = await res.json();

		const station = data.stations?.find((s) => s.id === parseInt(params.id));

		if (!station) {
			return { status: 404, error: new Error('Station not found') };
		}

		return { station };
	} catch (error) {
		console.error(error);
		return { status: 500, error };
	}
};
